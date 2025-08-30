package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"go.uber.org/zap"

	"airline-booking/internal/db"
	"airline-booking/internal/models"
)

type TicketRepository struct {
	db     *db.Database
	logger *zap.Logger
}

func NewTicketRepository(database *db.Database, logger *zap.Logger) *TicketRepository {
	return &TicketRepository{
		db:     database,
		logger: logger,
	}
}

// CreateTicket creates a new ticket in a transaction
func (r *TicketRepository) CreateTicket(ctx context.Context, tx *sql.Tx, ticket models.Ticket) (*models.Ticket, error) {
	queries := r.db.WithTx(tx)
	
	// Generate PNR code
	pnrCode := r.generatePNRCode()
	
	ticketID, err := queries.CreateTicket(ctx, db.CreateTicketParams{
		FlightID:    ticket.FlightID,
		SeatNo:      ticket.SeatNo,
		UserID:      ticket.UserID,
		PriceAmount: ticket.PriceAmount,
		Currency:    ticket.Currency,
		PnrCode:     pnrCode,
		PaymentRef:  ticket.PaymentRef,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	
	// Get the created ticket
	createdTicket, err := queries.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created ticket: %w", err)
	}
	
	result := &models.Ticket{
		ID:          createdTicket.ID,
		FlightID:    createdTicket.FlightID,
		SeatNo:      createdTicket.SeatNo,
		UserID:      createdTicket.UserID,
		PriceAmount: createdTicket.PriceAmount,
		Currency:    createdTicket.Currency,
		IssuedAt:    createdTicket.IssuedAt,
		PNRCode:     createdTicket.PnrCode,
		PaymentRef:  createdTicket.PaymentRef,
		CreatedAt:   createdTicket.CreatedAt,
	}
	
	r.logger.Info("Ticket created successfully",
		zap.Int64("ticket_id", result.ID),
		zap.String("pnr_code", result.PNRCode),
		zap.Int64("flight_id", result.FlightID),
		zap.String("seat_no", result.SeatNo))
	
	return result, nil
}

// GetTicketByPNR retrieves a ticket by PNR code
func (r *TicketRepository) GetTicketByPNR(ctx context.Context, pnrCode string) (*models.Ticket, error) {
	ticket, err := r.db.Queries.GetTicketByPNR(ctx, pnrCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ticket by PNR: %w", err)
	}
	
	return &models.Ticket{
		ID:          ticket.ID,
		FlightID:    ticket.FlightID,
		SeatNo:      ticket.SeatNo,
		UserID:      ticket.UserID,
		PriceAmount: ticket.PriceAmount,
		Currency:    ticket.Currency,
		IssuedAt:    ticket.IssuedAt,
		PNRCode:     ticket.PnrCode,
		PaymentRef:  ticket.PaymentRef,
		CreatedAt:   ticket.CreatedAt,
	}, nil
}

// GetTicketByFlightSeat checks if a ticket exists for a flight/seat combination
func (r *TicketRepository) GetTicketByFlightSeat(ctx context.Context, flightID int64, seatNo string) (*models.Ticket, error) {
	ticket, err := r.db.Queries.GetTicketByFlightSeat(ctx, db.GetTicketByFlightSeatParams{
		FlightID: flightID,
		SeatNo:   seatNo,
	})
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ticket by flight/seat: %w", err)
	}
	
	return &models.Ticket{
		ID:          ticket.ID,
		FlightID:    ticket.FlightID,
		SeatNo:      ticket.SeatNo,
		UserID:      ticket.UserID,
		PriceAmount: ticket.PriceAmount,
		Currency:    ticket.Currency,
		IssuedAt:    ticket.IssuedAt,
		PNRCode:     ticket.PnrCode,
		PaymentRef:  ticket.PaymentRef,
		CreatedAt:   ticket.CreatedAt,
	}, nil
}

// ListUserTickets retrieves all tickets for a user
func (r *TicketRepository) ListUserTickets(ctx context.Context, userID string) ([]models.Ticket, error) {
	tickets, err := r.db.Queries.ListUserTickets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user tickets: %w", err)
	}
	
	result := make([]models.Ticket, len(tickets))
	for i, ticket := range tickets {
		result[i] = models.Ticket{
			ID:          ticket.ID,
			FlightID:    ticket.FlightID,
			SeatNo:      ticket.SeatNo,
			UserID:      ticket.UserID,
			PriceAmount: ticket.PriceAmount,
			Currency:    ticket.Currency,
			IssuedAt:    ticket.IssuedAt,
			PNRCode:     ticket.PnrCode,
			PaymentRef:  ticket.PaymentRef,
			CreatedAt:   ticket.CreatedAt,
		}
	}
	
	return result, nil
}

// generatePNRCode generates a random 6-character PNR code
func (r *TicketRepository) generatePNRCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
