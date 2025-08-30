package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"airline-booking/internal/db"
	"airline-booking/internal/models"
)

type SeatRepository struct {
	db     *db.Database
	logger *zap.Logger
}

func NewSeatRepository(database *db.Database, logger *zap.Logger) *SeatRepository {
	return &SeatRepository{
		db:     database,
		logger: logger,
	}
}

// CreateHold attempts to create or update a seat hold using compare-and-set logic
func (r *SeatRepository) CreateHold(ctx context.Context, flightID int64, seatNo, holderID string, expiresAt time.Time) error {
	// First try to insert a new lock
	err := r.db.Queries.CreateSeatLock(ctx, db.CreateSeatLockParams{
		FlightID:  flightID,
		SeatNo:    seatNo,
		HolderID:  holderID,
		ExpiresAt: &expiresAt,
	})
	
	if err != nil {
		// If insert fails due to duplicate key, try to update with CAS logic
		rowsAffected, updateErr := r.db.Queries.UpdateSeatLock(ctx, db.UpdateSeatLockParams{
			HolderID:   holderID,
			ExpiresAt:  &expiresAt,
			FlightID:   flightID,
			SeatNo:     seatNo,
			HolderID_2: holderID, // For the OR condition in WHERE clause
		})
		
		if updateErr != nil {
			return fmt.Errorf("failed to update seat lock: %w", updateErr)
		}
		
		if rowsAffected == 0 {
			return fmt.Errorf("seat is already held by another user")
		}
	}
	
	r.logger.Info("Seat hold created/updated successfully",
		zap.Int64("flight_id", flightID),
		zap.String("seat_no", seatNo),
		zap.String("holder_id", holderID),
		zap.Time("expires_at", expiresAt))
	
	return nil
}

// GetSeatLock retrieves a seat lock
func (r *SeatRepository) GetSeatLock(ctx context.Context, flightID int64, seatNo string) (*models.SeatLock, error) {
	lock, err := r.db.Queries.GetSeatLock(ctx, db.GetSeatLockParams{
		FlightID: flightID,
		SeatNo:   seatNo,
	})
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get seat lock: %w", err)
	}
	
	return &models.SeatLock{
		ID:        lock.ID,
		FlightID:  lock.FlightID,
		SeatNo:    lock.SeatNo,
		HolderID:  lock.HolderID,
		ExpiresAt: lock.ExpiresAt,
		CreatedAt: lock.CreatedAt,
		UpdatedAt: lock.UpdatedAt,
	}, nil
}

// ConfirmHold converts a hold to a permanent ticket lock
func (r *SeatRepository) ConfirmHold(ctx context.Context, flightID int64, seatNo, holderID string) error {
	rowsAffected, err := r.db.Queries.ConfirmSeatLock(ctx, db.ConfirmSeatLockParams{
		FlightID: flightID,
		SeatNo:   seatNo,
		HolderID: holderID,
	})
	
	if err != nil {
		return fmt.Errorf("failed to confirm seat lock: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("no valid hold found to confirm")
	}
	
	r.logger.Info("Seat hold confirmed successfully",
		zap.Int64("flight_id", flightID),
		zap.String("seat_no", seatNo),
		zap.String("holder_id", holderID))
	
	return nil
}

// ReleaseHold releases a seat hold
func (r *SeatRepository) ReleaseHold(ctx context.Context, flightID int64, seatNo, holderID string) error {
	err := r.db.Queries.ReleaseSeatLock(ctx, db.ReleaseSeatLockParams{
		FlightID: flightID,
		SeatNo:   seatNo,
		HolderID: holderID,
	})
	
	if err != nil {
		return fmt.Errorf("failed to release seat lock: %w", err)
	}
	
	r.logger.Info("Seat hold released successfully",
		zap.Int64("flight_id", flightID),
		zap.String("seat_no", seatNo),
		zap.String("holder_id", holderID))
	
	return nil
}

// CleanupExpiredHolds removes all expired holds
func (r *SeatRepository) CleanupExpiredHolds(ctx context.Context) error {
	err := r.db.Queries.CleanupExpiredLocks(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired locks: %w", err)
	}
	
	r.logger.Debug("Expired holds cleaned up successfully")
	return nil
}

// GetFlightSeatAvailability returns seat availability for a flight
func (r *SeatRepository) GetFlightSeatAvailability(ctx context.Context, flightID int64) ([]models.SeatAvailability, error) {
	seats, err := r.db.Queries.ListSeats(ctx, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to list seats: %w", err)
	}
	
	locks, err := r.db.Queries.ListFlightSeatLocks(ctx, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to list seat locks: %w", err)
	}
	
	tickets, err := r.db.Queries.ListFlightTickets(ctx, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}
	
	// Create maps for quick lookup
	lockMap := make(map[string]*db.SeatLock)
	for _, lock := range locks {
		lockMap[lock.SeatNo] = &lock
	}
	
	ticketMap := make(map[string]bool)
	for _, ticket := range tickets {
		ticketMap[ticket.SeatNo] = true
	}
	
	// Build availability list
	availability := make([]models.SeatAvailability, len(seats))
	for i, seat := range seats {
		seatAvail := models.SeatAvailability{
			SeatNo: seat.SeatNo,
			Class:  seat.Class,
			Price:  29900, // Base price in cents ($299.00)
		}
		
		// Check if sold
		if ticketMap[seat.SeatNo] {
			seatAvail.Status = models.SeatStatusSold
		} else if lock, exists := lockMap[seat.SeatNo]; exists {
			// Check if lock is expired
			if lock.ExpiresAt != nil && lock.ExpiresAt.Before(time.Now()) {
				seatAvail.Status = models.SeatStatusAvailable
			} else {
				seatAvail.Status = models.SeatStatusHeld
				seatAvail.ExpiresAt = lock.ExpiresAt
			}
		} else {
			seatAvail.Status = models.SeatStatusAvailable
		}
		
		availability[i] = seatAvail
	}
	
	return availability, nil
}

// GetHold is an alias for GetSeatLock for consistency with the booking service
func (r *SeatRepository) GetHold(ctx context.Context, flightID int64, seatNo string) (*models.SeatLock, error) {
	return r.GetSeatLock(ctx, flightID, seatNo)
}
