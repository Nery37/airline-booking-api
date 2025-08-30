package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"airline-booking/internal/config"
	"airline-booking/internal/db"
	"airline-booking/internal/es"
	"airline-booking/internal/models"
	"airline-booking/internal/repository"
)

type BookingService struct {
	seatRepo   *repository.SeatRepository
	ticketRepo *repository.TicketRepository
	flightRepo *repository.FlightRepository
	esClient   *es.Client
	db         *db.Database
	config     *config.Config
	logger     *zap.Logger
}

func NewBookingService(
	seatRepo *repository.SeatRepository,
	ticketRepo *repository.TicketRepository,
	flightRepo *repository.FlightRepository,
	esClient *es.Client,
	database *db.Database,
	cfg *config.Config,
	logger *zap.Logger,
) *BookingService {
	return &BookingService{
		seatRepo:   seatRepo,
		ticketRepo: ticketRepo,
		flightRepo: flightRepo,
		esClient:   esClient,
		db:         database,
		config:     cfg,
		logger:     logger,
	}
}

// CreateHold creates a seat hold with idempotency support
func (s *BookingService) CreateHold(ctx context.Context, req models.CreateHoldRequest, holderID, idempotencyKey string) (*models.CreateHoldResponse, error) {
	// Check idempotency if key provided
	if idempotencyKey != "" {
		if response, err := s.checkIdempotency(ctx, idempotencyKey, "POST /holds", holderID); err == nil && response != nil {
			return response.(*models.CreateHoldResponse), nil
		}
	}
	
	// Validate flight exists
	flight, err := s.flightRepo.GetFlight(ctx, req.FlightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight: %w", err)
	}
	if flight == nil {
		return nil, fmt.Errorf("flight not found")
	}
	
	// Check if seat is already ticketed
	existingTicket, err := s.ticketRepo.GetTicketByFlightSeat(ctx, req.FlightID, req.SeatNo)
	s.logger.Info("Checked existing ticket", 
		zap.Int64("flight_id", req.FlightID), 
		zap.String("seat_no", req.SeatNo),
		zap.Bool("ticket_exists", existingTicket != nil),
		zap.Error(err))
	if err != nil {
		return nil, fmt.Errorf("failed to check existing ticket: %w", err)
	}
	if existingTicket != nil {
		return nil, fmt.Errorf("seat is already sold")
	}
	
	// Calculate expiration time
	expiresAt := time.Now().UTC().Add(s.config.Hold.TTL)
	
	// Attempt to create hold
	err = s.seatRepo.CreateHold(ctx, req.FlightID, req.SeatNo, holderID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create hold: %w", err)
	}

	// Get the created hold for indexing
	hold, err := s.seatRepo.GetHold(ctx, req.FlightID, req.SeatNo)
	if err != nil {
		s.logger.Warn("Failed to get hold for indexing", zap.Error(err))
	} else {
		// Index hold in Elasticsearch
		holdDoc := es.HoldDocument{
			ID:        hold.ID,
			FlightID:  hold.FlightID,
			SeatNo:    hold.SeatNo,
			HolderID:  hold.HolderID,
			ExpiresAt: hold.ExpiresAt,
			CreatedAt: hold.CreatedAt,
			UpdatedAt: hold.UpdatedAt,
			Status:    "active",
		}
		
		if err := s.esClient.IndexHold(ctx, holdDoc); err != nil {
			s.logger.Error("Failed to index hold in Elasticsearch", 
				zap.Error(err),
				zap.Int64("hold_id", hold.ID))
			// Don't fail the request if ES indexing fails
		}
	}
	
	response := &models.CreateHoldResponse{
		FlightID:  req.FlightID,
		SeatNo:    req.SeatNo,
		HolderID:  holderID,
		ExpiresAt: expiresAt,
	}
	
	// Store idempotency key if provided
	if idempotencyKey != "" {
		if err := s.storeIdempotency(ctx, idempotencyKey, "POST /holds", holderID, response); err != nil {
			s.logger.Warn("Failed to store idempotency key", zap.Error(err))
		}
	}
	
	s.logger.Info("Hold created successfully",
		zap.Int64("flight_id", req.FlightID),
		zap.String("seat_no", req.SeatNo),
		zap.String("holder_id", holderID))
	
	return response, nil
}

// ConfirmTicket confirms a hold and creates a ticket
func (s *BookingService) ConfirmTicket(ctx context.Context, req models.ConfirmTicketRequest, userID, idempotencyKey string) (*models.ConfirmTicketResponse, error) {
	// Check idempotency if key provided
	if idempotencyKey != "" {
		if response, err := s.checkIdempotency(ctx, idempotencyKey, "POST /tickets/confirm", userID); err == nil && response != nil {
			return response.(*models.ConfirmTicketResponse), nil
		}
	}
	
	// Validate flight exists
	flight, err := s.flightRepo.GetFlight(ctx, req.FlightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight: %w", err)
	}
	if flight == nil {
		return nil, fmt.Errorf("flight not found")
	}
	
	// Start transaction
	tx, err := s.db.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Confirm the hold (this will set expires_at to NULL)
	err = s.seatRepo.ConfirmHold(ctx, req.FlightID, req.SeatNo, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm hold: %w", err)
	}
	
	// Create ticket
	ticket := models.Ticket{
		FlightID:    req.FlightID,
		SeatNo:      req.SeatNo,
		UserID:      userID,
		PriceAmount: 29900, // $299.00 in cents
		Currency:    "USD",
		PaymentRef:  req.PaymentRef,
	}
	
	createdTicket, err := s.ticketRepo.CreateTicket(ctx, tx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Index ticket in Elasticsearch
	ticketDoc := es.TicketDocument{
		ID:          createdTicket.ID,
		FlightID:    createdTicket.FlightID,
		SeatNo:      createdTicket.SeatNo,
		UserID:      createdTicket.UserID,
		PriceAmount: createdTicket.PriceAmount,
		Currency:    createdTicket.Currency,
		IssuedAt:    createdTicket.IssuedAt,
		PnrCode:     createdTicket.PNRCode,
		PaymentRef:  createdTicket.PaymentRef,
		CreatedAt:   createdTicket.CreatedAt,
		Status:      "confirmed",
	}
	
	if err := s.esClient.IndexTicket(ctx, ticketDoc); err != nil {
		s.logger.Error("Failed to index ticket in Elasticsearch", 
			zap.Error(err),
			zap.Int64("ticket_id", createdTicket.ID))
		// Don't fail the request if ES indexing fails
	}

	// Update hold status to confirmed in Elasticsearch (if we can find it)
	if hold, err := s.seatRepo.GetHold(ctx, req.FlightID, req.SeatNo); err == nil {
		if err := s.esClient.UpdateHoldStatus(ctx, hold.ID, "confirmed"); err != nil {
			s.logger.Warn("Failed to update hold status in Elasticsearch", zap.Error(err))
		}
	}
	
	response := &models.ConfirmTicketResponse{
		TicketID:   createdTicket.ID,
		FlightID:   createdTicket.FlightID,
		SeatNo:     createdTicket.SeatNo,
		PNRCode:    createdTicket.PNRCode,
		PaymentRef: createdTicket.PaymentRef,
	}
	
	// Store idempotency key if provided
	if idempotencyKey != "" {
		if err := s.storeIdempotency(ctx, idempotencyKey, "POST /tickets/confirm", userID, response); err != nil {
			s.logger.Warn("Failed to store idempotency key", zap.Error(err))
		}
	}
	
	s.logger.Info("Ticket confirmed successfully",
		zap.Int64("ticket_id", createdTicket.ID),
		zap.String("pnr_code", createdTicket.PNRCode),
		zap.Int64("flight_id", req.FlightID),
		zap.String("seat_no", req.SeatNo))
	
	return response, nil
}

// ReleaseHold releases a hold for a specific user
func (s *BookingService) ReleaseHold(ctx context.Context, flightID int64, seatNo, holderID string) error {
	// Get hold before releasing to get the ID for ES deletion
	hold, err := s.seatRepo.GetHold(ctx, flightID, seatNo)
	if err != nil {
		s.logger.Warn("Failed to get hold for ES deletion", zap.Error(err))
	}

	err = s.seatRepo.ReleaseHold(ctx, flightID, seatNo, holderID)
	if err != nil {
		return fmt.Errorf("failed to release hold: %w", err)
	}
	
	// Delete hold from Elasticsearch if we got the hold info
	if hold != nil {
		if err := s.esClient.DeleteHold(ctx, hold.ID); err != nil {
			s.logger.Warn("Failed to delete hold from Elasticsearch", 
				zap.Error(err),
				zap.Int64("hold_id", hold.ID))
			// Don't fail the request if ES deletion fails
		}
	}
	
	s.logger.Info("Hold released successfully",
		zap.Int64("flight_id", flightID),
		zap.String("seat_no", seatNo),
		zap.String("holder_id", holderID))
	
	return nil
}

// GetFlightSeatAvailability returns seat availability for a flight
func (s *BookingService) GetFlightSeatAvailability(ctx context.Context, flightID int64) ([]models.SeatAvailability, error) {
	availability, err := s.seatRepo.GetFlightSeatAvailability(ctx, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seat availability: %w", err)
	}
	
	return availability, nil
}

// SearchFlights searches for flights using Elasticsearch
func (s *BookingService) SearchFlights(ctx context.Context, req models.FlightSearchRequest) (*models.FlightSearchResponse, error) {
	// Search in Elasticsearch
	esResponse, err := s.esClient.SearchFlights(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search flights in elasticsearch: %w", err)
	}
	
	// For each flight, get available seat count by checking locks and tickets
	for i := range esResponse.Flights {
		flight := &esResponse.Flights[i]
		
		availability, err := s.seatRepo.GetFlightSeatAvailability(ctx, flight.ID)
		if err != nil {
			s.logger.Warn("Failed to get seat availability for flight",
				zap.Int64("flight_id", flight.ID),
				zap.Error(err))
			continue
		}
		
		// Count available seats
		availableCount := 0
		for _, seat := range availability {
			if seat.Status == models.SeatStatusAvailable {
				availableCount++
			}
		}
		
		flight.AvailableSeats = availableCount
	}
	
	return esResponse, nil
}

// CleanupExpiredHolds removes expired holds
func (s *BookingService) CleanupExpiredHolds(ctx context.Context) error {
	return s.seatRepo.CleanupExpiredHolds(ctx)
}

// Helper methods for idempotency
func (s *BookingService) checkIdempotency(ctx context.Context, requestID, route, userID string) (interface{}, error) {
	_, err := s.db.Queries.GetIdempotencyKey(ctx, db.GetIdempotencyKeyParams{
		RequestID: requestID,
		Route:     route,
	})
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, proceed with request
		}
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}
	
	// TODO: In a real implementation, we would deserialize the response
	// For this POC, we'll return a simple response
	s.logger.Info("Idempotent request detected", zap.String("request_id", requestID))
	return nil, nil
}

func (s *BookingService) storeIdempotency(ctx context.Context, requestID, route, userID string, response interface{}) error {
	// Create a hash of the response for storage
	responseHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%+v", response))))
	
	return s.db.Queries.CreateIdempotencyKey(ctx, db.CreateIdempotencyKeyParams{
		RequestID:    requestID,
		Route:        route,
		UserID:       userID,
		ResponseHash: responseHash,
	})
}

// CreateFlight creates a new flight and indexes it in Elasticsearch
func (s *BookingService) CreateFlight(ctx context.Context, req models.CreateFlightRequest) (*models.CreateFlightResponse, error) {
	s.logger.Info("CreateFlight called", 
		zap.String("origin", req.Origin),
		zap.String("destination", req.Destination),
		zap.String("airline", req.Airline))

	// Parse times
	departureTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		s.logger.Error("Failed to parse departure_time", zap.Error(err), zap.String("departure_time", req.DepartureTime))
		return nil, fmt.Errorf("invalid departure_time format: %w", err)
	}
	
	arrivalTime, err := time.Parse(time.RFC3339, req.ArrivalTime)
	if err != nil {
		s.logger.Error("Failed to parse arrival_time", zap.Error(err), zap.String("arrival_time", req.ArrivalTime))
		return nil, fmt.Errorf("invalid arrival_time format: %w", err)
	}
	
	// Validate business logic
	if arrivalTime.Before(departureTime) {
		s.logger.Error("Arrival time before departure time")
		return nil, fmt.Errorf("arrival time cannot be before departure time")
	}
	
	// Create flight in database
	flight := models.Flight{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureTime: departureTime,
		ArrivalTime:   arrivalTime,
		Airline:       req.Airline,
		Aircraft:      req.Aircraft,
		FareClass:     req.FareClass,
	}
	
	s.logger.Info("Calling flightRepo.CreateFlight")
	log.Printf("DEBUG Service - About to call repository with origin: %s, destination: %s", flight.Origin, flight.Destination)
	createdFlight, err := s.flightRepo.CreateFlight(ctx, flight)
	if err != nil {
		s.logger.Error("Failed to create flight in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create flight in database: %w", err)
	}
	
	log.Printf("DEBUG Service - Repository returned flight ID: %d", createdFlight.ID)
	s.logger.Info("Flight created in repository", zap.Int64("flight_id", createdFlight.ID))
	
	// Create seats if seat configuration is provided
	seatsCreated := 0
	if req.SeatConfig != nil {
		seats := s.generateSeats(*req.SeatConfig, req.BasePrice)
		err = s.flightRepo.CreateSeats(ctx, createdFlight.ID, seats)
		if err != nil {
			s.logger.Warn("Failed to create seats for flight", 
				zap.Int64("flight_id", createdFlight.ID), 
				zap.Error(err))
		} else {
			seatsCreated = len(seats)
		}
	}
	
	// Index in Elasticsearch
	esDoc := es.FlightDocument{
		ID:            createdFlight.ID,
		Origin:        createdFlight.Origin,
		Destination:   createdFlight.Destination,
		DepartureTime: createdFlight.DepartureTime,
		ArrivalTime:   createdFlight.ArrivalTime,
		Airline:       createdFlight.Airline,
		Aircraft:      createdFlight.Aircraft,
		FareClass:     createdFlight.FareClass,
		BasePrice:     req.BasePrice,
	}
	
	if err := s.esClient.IndexFlight(ctx, esDoc); err != nil {
		s.logger.Error("Failed to index flight in Elasticsearch", 
			zap.Int64("flight_id", createdFlight.ID), 
			zap.Error(err))
		// Don't fail the request, just log the error
	}
	
	s.logger.Info("Flight created successfully", 
		zap.Int64("flight_id", createdFlight.ID),
		zap.String("route", fmt.Sprintf("%s -> %s", req.Origin, req.Destination)),
		zap.Int("seats_created", seatsCreated))
	
	return &models.CreateFlightResponse{
		ID:            createdFlight.ID,
		Origin:        createdFlight.Origin,
		Destination:   createdFlight.Destination,
		DepartureTime: createdFlight.DepartureTime.Format(time.RFC3339),
		ArrivalTime:   createdFlight.ArrivalTime.Format(time.RFC3339),
		Airline:       createdFlight.Airline,
		Aircraft:      createdFlight.Aircraft,
		FareClass:     createdFlight.FareClass,
		BasePrice:     req.BasePrice,
		SeatsCreated:  seatsCreated,
		CreatedAt:     createdFlight.CreatedAt.Format(time.RFC3339),
	}, nil
}

// generateSeats creates seat configuration based on the provided configuration
func (s *BookingService) generateSeats(config models.SeatConfiguration, basePrice float64) []models.Seat {
	var seats []models.Seat
	seatLetters := []string{"A", "B", "C", "D", "E", "F"}
	
	if config.SeatsPerRow > len(seatLetters) {
		config.SeatsPerRow = len(seatLetters)
	}
	
	currentRow := 1
	
	// First class seats
	for i := 0; i < config.FirstClassRows; i++ {
		for j := 0; j < config.SeatsPerRow; j++ {
			seats = append(seats, models.Seat{
				SeatNo: fmt.Sprintf("%d%s", currentRow, seatLetters[j]),
				Class:  "first",
			})
		}
		currentRow++
	}
	
	// Business class seats
	for i := 0; i < config.BusinessRows; i++ {
		for j := 0; j < config.SeatsPerRow; j++ {
			seats = append(seats, models.Seat{
				SeatNo: fmt.Sprintf("%d%s", currentRow, seatLetters[j]),
				Class:  "business",
			})
		}
		currentRow++
	}
	
	// Economy class seats
	for i := 0; i < config.EconomyRows; i++ {
		for j := 0; j < config.SeatsPerRow; j++ {
			seats = append(seats, models.Seat{
				SeatNo: fmt.Sprintf("%d%s", currentRow, seatLetters[j]),
				Class:  "economy",
			})
		}
		currentRow++
	}
	
	return seats
}
