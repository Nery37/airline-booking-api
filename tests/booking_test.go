package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"airline-booking/internal/config"
	"airline-booking/internal/db"
	"airline-booking/internal/es"
	"airline-booking/internal/models"
	"airline-booking/internal/repository"
	"airline-booking/internal/service"
)

func TestConcurrentHoldCreation(t *testing.T) {
	// Setup test environment
	cfg, err := config.Load()
	require.NoError(t, err)

	// Use test database
	cfg.Database.Name = "airline_booking_test"

	logger, _ := zap.NewDevelopment()
	database, err := db.NewDatabase(&cfg.Database, logger)
	require.NoError(t, err)
	defer database.Close()

	// Run migrations
	err = database.RunMigrations("../migrations")
	require.NoError(t, err)

	// Setup Elasticsearch mock or skip ES tests
	esClient, err := es.NewClient(&cfg.Elasticsearch, logger)
	if err != nil {
		t.Skip("Elasticsearch not available, skipping test")
	}

	// Initialize repositories and service
	seatRepo := repository.NewSeatRepository(database, logger)
	ticketRepo := repository.NewTicketRepository(database, logger)
	flightRepo := repository.NewFlightRepository(database, logger)

	bookingService := service.NewBookingService(
		seatRepo,
		ticketRepo,
		flightRepo,
		esClient,
		database,
		cfg,
		logger,
	)

	ctx := context.Background()

	// Create a test flight with seats
	flight := models.Flight{
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureTime: time.Now().Add(24 * time.Hour),
		ArrivalTime:   time.Now().Add(29 * time.Hour),
		Airline:       "AA",
		Aircraft:      "Boeing 737",
		FareClass:     "economy",
	}

	createdFlight, err := flightRepo.CreateFlight(ctx, flight)
	require.NoError(t, err)

	// Create test seats
	seats := []models.Seat{
		{FlightID: createdFlight.ID, SeatNo: "12A", Class: "economy"},
		{FlightID: createdFlight.ID, SeatNo: "12B", Class: "economy"},
	}
	err = flightRepo.CreateSeats(ctx, createdFlight.ID, seats)
	require.NoError(t, err)

	t.Run("ConcurrentHoldsSameSeat", func(t *testing.T) {
		numGoroutines := 10
		var wg sync.WaitGroup
		results := make([]error, numGoroutines)

		// All goroutines try to hold the same seat
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				req := models.CreateHoldRequest{
					FlightID: createdFlight.ID,
					SeatNo:   "12A",
				}

				holderID := fmt.Sprintf("user_%d", index)
				_, err := bookingService.CreateHold(ctx, req, holderID, "")
				results[index] = err
			}(i)
		}

		wg.Wait()

		// Count successful holds
		successCount := 0
		conflictCount := 0

		for _, err := range results {
			if err == nil {
				successCount++
			} else if err.Error() == "seat is already held by another user" {
				conflictCount++
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		}

		// Only one should succeed
		assert.Equal(t, 1, successCount, "Expected exactly one successful hold")
		assert.Equal(t, numGoroutines-1, conflictCount, "Expected conflicts for all other attempts")
	})

	t.Run("ConcurrentHoldsDifferentSeats", func(t *testing.T) {
		// Clean up previous test
		err := seatRepo.CleanupExpiredHolds(ctx)
		require.NoError(t, err)

		numGoroutines := 2
		var wg sync.WaitGroup
		results := make([]error, numGoroutines)
		seats := []string{"12A", "12B"}

		// Each goroutine tries to hold a different seat
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				req := models.CreateHoldRequest{
					FlightID: createdFlight.ID,
					SeatNo:   seats[index],
				}

				holderID := fmt.Sprintf("user_%d", index)
				_, err := bookingService.CreateHold(ctx, req, holderID, "")
				results[index] = err
			}(i)
		}

		wg.Wait()

		// Both should succeed
		for i, err := range results {
			assert.NoError(t, err, "Hold %d should succeed", i)
		}
	})
}

func TestHoldExpiration(t *testing.T) {
	// Setup test environment with short TTL
	cfg, err := config.Load()
	require.NoError(t, err)

	// Use short TTL for testing
	cfg.Hold.TTL = 2 * time.Second
	cfg.Database.Name = "airline_booking_test"

	logger, _ := zap.NewDevelopment()
	database, err := db.NewDatabase(&cfg.Database, logger)
	require.NoError(t, err)
	defer database.Close()

	// Initialize repositories and service
	seatRepo := repository.NewSeatRepository(database, logger)
	ticketRepo := repository.NewTicketRepository(database, logger)
	flightRepo := repository.NewFlightRepository(database, logger)

	esClient, err := es.NewClient(&cfg.Elasticsearch, logger)
	if err != nil {
		t.Skip("Elasticsearch not available, skipping test")
	}

	bookingService := service.NewBookingService(
		seatRepo,
		ticketRepo,
		flightRepo,
		esClient,
		database,
		cfg,
		logger,
	)

	ctx := context.Background()

	// Create a test flight with seats
	flight := models.Flight{
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureTime: time.Now().Add(24 * time.Hour),
		ArrivalTime:   time.Now().Add(29 * time.Hour),
		Airline:       "AA",
		Aircraft:      "Boeing 737",
		FareClass:     "economy",
	}

	createdFlight, err := flightRepo.CreateFlight(ctx, flight)
	require.NoError(t, err)

	seats := []models.Seat{
		{FlightID: createdFlight.ID, SeatNo: "15A", Class: "economy"},
	}
	err = flightRepo.CreateSeats(ctx, createdFlight.ID, seats)
	require.NoError(t, err)

	// Create hold
	req := models.CreateHoldRequest{
		FlightID: createdFlight.ID,
		SeatNo:   "15A",
	}

	holderID := "test_user"
	holdResponse, err := bookingService.CreateHold(ctx, req, holderID, "")
	require.NoError(t, err)
	assert.True(t, holdResponse.ExpiresAt.After(time.Now()))

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Try to create another hold - should succeed after expiration
	anotherHolderID := "another_user"
	_, err = bookingService.CreateHold(ctx, req, anotherHolderID, "")
	assert.NoError(t, err, "Should be able to create hold after expiration")
}

func TestTicketConfirmation(t *testing.T) {
	// Setup test environment
	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.Database.Name = "airline_booking_test"

	logger, _ := zap.NewDevelopment()
	database, err := db.NewDatabase(&cfg.Database, logger)
	require.NoError(t, err)
	defer database.Close()

	// Initialize repositories and service
	seatRepo := repository.NewSeatRepository(database, logger)
	ticketRepo := repository.NewTicketRepository(database, logger)
	flightRepo := repository.NewFlightRepository(database, logger)

	esClient, err := es.NewClient(&cfg.Elasticsearch, logger)
	if err != nil {
		t.Skip("Elasticsearch not available, skipping test")
	}

	bookingService := service.NewBookingService(
		seatRepo,
		ticketRepo,
		flightRepo,
		esClient,
		database,
		cfg,
		logger,
	)

	ctx := context.Background()

	// Create a test flight with seats
	flight := models.Flight{
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureTime: time.Now().Add(24 * time.Hour),
		ArrivalTime:   time.Now().Add(29 * time.Hour),
		Airline:       "AA",
		Aircraft:      "Boeing 737",
		FareClass:     "economy",
	}

	createdFlight, err := flightRepo.CreateFlight(ctx, flight)
	require.NoError(t, err)

	seats := []models.Seat{
		{FlightID: createdFlight.ID, SeatNo: "20A", Class: "economy"},
	}
	err = flightRepo.CreateSeats(ctx, createdFlight.ID, seats)
	require.NoError(t, err)

	// Create hold first
	holdReq := models.CreateHoldRequest{
		FlightID: createdFlight.ID,
		SeatNo:   "20A",
	}

	userID := "test_user"
	_, err = bookingService.CreateHold(ctx, holdReq, userID, "")
	require.NoError(t, err)

	// Confirm ticket
	confirmReq := models.ConfirmTicketRequest{
		FlightID:   createdFlight.ID,
		SeatNo:     "20A",
		PaymentRef: "payment_123",
	}

	ticketResponse, err := bookingService.ConfirmTicket(ctx, confirmReq, userID, "")
	require.NoError(t, err)
	assert.NotEmpty(t, ticketResponse.PNRCode)
	assert.Equal(t, "payment_123", ticketResponse.PaymentRef)

	// Try to create another hold on the same seat - should fail
	anotherUserID := "another_user"
	_, err = bookingService.CreateHold(ctx, holdReq, anotherUserID, "")
	assert.Error(t, err, "Should not be able to hold a sold seat")
}
