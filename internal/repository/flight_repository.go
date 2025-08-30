package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"go.uber.org/zap"

	"airline-booking/internal/db"
	"airline-booking/internal/models"
)

type FlightRepository struct {
	db     *db.Database
	logger *zap.Logger
}

func NewFlightRepository(database *db.Database, logger *zap.Logger) *FlightRepository {
	return &FlightRepository{
		db:     database,
		logger: logger,
	}
}

// GetFlight retrieves a flight by ID
func (r *FlightRepository) GetFlight(ctx context.Context, id int64) (*models.Flight, error) {
	flight, err := r.db.Queries.GetFlight(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get flight: %w", err)
	}
	
	return &models.Flight{
		ID:            flight.ID,
		Origin:        flight.Origin,
		Destination:   flight.Destination,
		DepartureTime: flight.DepartureTime,
		ArrivalTime:   flight.ArrivalTime,
		Airline:       flight.Airline,
		Aircraft:      flight.Aircraft,
		FareClass:     flight.FareClass,
		CreatedAt:     flight.CreatedAt,
		UpdatedAt:     flight.UpdatedAt,
	}, nil
}

// CreateFlight creates a new flight
func (r *FlightRepository) CreateFlight(ctx context.Context, flight models.Flight) (*models.Flight, error) {
	log.Printf("DEBUG Repository.CreateFlight - Called with origin: %s, destination: %s", flight.Origin, flight.Destination)
	flightID, err := r.db.Queries.CreateFlight(ctx, db.CreateFlightParams{
		Origin:        flight.Origin,
		Destination:   flight.Destination,
		DepartureTime: flight.DepartureTime,
		ArrivalTime:   flight.ArrivalTime,
		Airline:       flight.Airline,
		Aircraft:      flight.Aircraft,
		FareClass:     flight.FareClass,
	})
	
	log.Printf("DEBUG Repository.CreateFlight - CreateFlight returned ID: %d, err: %v", flightID, err)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create flight: %w", err)
	}
	
	// Get the created flight
	result, err := r.GetFlight(ctx, flightID)
	log.Printf("DEBUG Repository.CreateFlight - GetFlight returned: %+v, err: %v", result, err)
	return result, err
}

// CreateSeats creates seats for a flight
func (r *FlightRepository) CreateSeats(ctx context.Context, flightID int64, seats []models.Seat) error {
	for _, seat := range seats {
		_, err := r.db.Queries.CreateSeat(ctx, db.CreateSeatParams{
			FlightID: flightID,
			SeatNo:   seat.SeatNo,
			Class:    seat.Class,
		})
		
		if err != nil {
			return fmt.Errorf("failed to create seat %s: %w", seat.SeatNo, err)
		}
	}
	
	r.logger.Info("Seats created successfully",
		zap.Int64("flight_id", flightID),
		zap.Int("count", len(seats)))
	
	return nil
}
