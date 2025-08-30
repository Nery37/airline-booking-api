package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"airline-booking/internal/config"
	"airline-booking/internal/db"
	"airline-booking/internal/models"
	"airline-booking/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Setup logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("Starting database seeding...")

	// Initialize database
	database, err := db.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Initialize repository
	flightRepo := repository.NewFlightRepository(database, logger)

	ctx := context.Background()

	// Create sample flights
	flights := getSampleFlights()
	
	for _, flight := range flights {
		createdFlight, err := flightRepo.CreateFlight(ctx, flight)
		if err != nil {
			logger.Error("Failed to create flight", zap.Error(err))
			continue
		}
		
		// Create seats for the flight
		seats := generateSeats(createdFlight.ID)
		if err := flightRepo.CreateSeats(ctx, createdFlight.ID, seats); err != nil {
			logger.Error("Failed to create seats", zap.Error(err))
		}
		
		logger.Info("Created flight with seats",
			zap.Int64("flight_id", createdFlight.ID),
			zap.String("route", fmt.Sprintf("%s-%s", flight.Origin, flight.Destination)),
			zap.Int("seats", len(seats)))
	}

	logger.Info("Database seeding completed successfully")
}

func getSampleFlights() []models.Flight {
	now := time.Now().UTC()
	
	return []models.Flight{
		{
			Origin:        "JFK",
			Destination:   "LAX",
			DepartureTime: now.Add(24 * time.Hour),
			ArrivalTime:   now.Add(24*time.Hour + 5*time.Hour),
			Airline:       "AA",
			Aircraft:      "Boeing 737",
			FareClass:     "economy",
		},
		{
			Origin:        "LAX",
			Destination:   "JFK",
			DepartureTime: now.Add(48 * time.Hour),
			ArrivalTime:   now.Add(48*time.Hour + 5*time.Hour),
			Airline:       "AA",
			Aircraft:      "Boeing 737",
			FareClass:     "economy",
		},
		{
			Origin:        "JFK",
			Destination:   "MIA",
			DepartureTime: now.Add(36 * time.Hour),
			ArrivalTime:   now.Add(36*time.Hour + 3*time.Hour),
			Airline:       "DL",
			Aircraft:      "Airbus A320",
			FareClass:     "business",
		},
		{
			Origin:        "MIA",
			Destination:   "JFK",
			DepartureTime: now.Add(60 * time.Hour),
			ArrivalTime:   now.Add(60*time.Hour + 3*time.Hour),
			Airline:       "DL",
			Aircraft:      "Airbus A320",
			FareClass:     "business",
		},
		{
			Origin:        "ORD",
			Destination:   "DFW",
			DepartureTime: now.Add(72 * time.Hour),
			ArrivalTime:   now.Add(72*time.Hour + 2*time.Hour),
			Airline:       "UA",
			Aircraft:      "Boeing 777",
			FareClass:     "economy",
		},
		{
			Origin:        "DFW",
			Destination:   "ORD",
			DepartureTime: now.Add(96 * time.Hour),
			ArrivalTime:   now.Add(96*time.Hour + 2*time.Hour),
			Airline:       "UA",
			Aircraft:      "Boeing 777",
			FareClass:     "first",
		},
	}
}

func generateSeats(flightID int64) []models.Seat {
	seats := make([]models.Seat, 0, 150)
	
	// First class (rows 1-3, A-F)
	for row := 1; row <= 3; row++ {
		for _, letter := range []string{"A", "B", "C", "D", "E", "F"} {
			seats = append(seats, models.Seat{
				FlightID: flightID,
				SeatNo:   fmt.Sprintf("%d%s", row, letter),
				Class:    "first",
			})
		}
	}
	
	// Business class (rows 4-10, A-F)
	for row := 4; row <= 10; row++ {
		for _, letter := range []string{"A", "B", "C", "D", "E", "F"} {
			seats = append(seats, models.Seat{
				FlightID: flightID,
				SeatNo:   fmt.Sprintf("%d%s", row, letter),
				Class:    "business",
			})
		}
	}
	
	// Economy class (rows 11-30, A-F)
	for row := 11; row <= 30; row++ {
		for _, letter := range []string{"A", "B", "C", "D", "E", "F"} {
			seats = append(seats, models.Seat{
				FlightID: flightID,
				SeatNo:   fmt.Sprintf("%d%s", row, letter),
				Class:    "economy",
			})
		}
	}
	
	return seats
}
