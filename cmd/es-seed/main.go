package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"airline-booking/internal/config"
	"airline-booking/internal/es"
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

	logger.Info("Starting Elasticsearch seeding...")

	// Initialize Elasticsearch client
	esClient, err := es.NewClient(&cfg.Elasticsearch, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Elasticsearch", zap.Error(err))
	}

	ctx := context.Background()

	// Create index if not exists
	if err := esClient.CreateIndex(ctx); err != nil {
		logger.Fatal("Failed to create index", zap.Error(err))
	}

	// Create sample flight documents
	flights := getSampleFlightDocuments()

	// Bulk index the flights
	if err := esClient.BulkIndexFlights(ctx, flights); err != nil {
		logger.Fatal("Failed to bulk index flights", zap.Error(err))
	}

	logger.Info("Elasticsearch seeding completed successfully",
		zap.Int("flights_indexed", len(flights)))
}

func getSampleFlightDocuments() []es.FlightDocument {
	now := time.Now().UTC()

	return []es.FlightDocument{
		{
			ID:            1,
			Origin:        "JFK",
			Destination:   "LAX",
			DepartureTime: now.Add(24 * time.Hour),
			ArrivalTime:   now.Add(24*time.Hour + 5*time.Hour),
			Airline:       "AA",
			Aircraft:      "Boeing 737",
			FareClass:     "economy",
			BasePrice:     29900, // $299.00
		},
		{
			ID:            2,
			Origin:        "LAX",
			Destination:   "JFK",
			DepartureTime: now.Add(48 * time.Hour),
			ArrivalTime:   now.Add(48*time.Hour + 5*time.Hour),
			Airline:       "AA",
			Aircraft:      "Boeing 737",
			FareClass:     "economy",
			BasePrice:     31900, // $319.00
		},
		{
			ID:            3,
			Origin:        "JFK",
			Destination:   "MIA",
			DepartureTime: now.Add(36 * time.Hour),
			ArrivalTime:   now.Add(36*time.Hour + 3*time.Hour),
			Airline:       "DL",
			Aircraft:      "Airbus A320",
			FareClass:     "business",
			BasePrice:     89900, // $899.00
		},
		{
			ID:            4,
			Origin:        "MIA",
			Destination:   "JFK",
			DepartureTime: now.Add(60 * time.Hour),
			ArrivalTime:   now.Add(60*time.Hour + 3*time.Hour),
			Airline:       "DL",
			Aircraft:      "Airbus A320",
			FareClass:     "business",
			BasePrice:     92900, // $929.00
		},
		{
			ID:            5,
			Origin:        "ORD",
			Destination:   "DFW",
			DepartureTime: now.Add(72 * time.Hour),
			ArrivalTime:   now.Add(72*time.Hour + 2*time.Hour),
			Airline:       "UA",
			Aircraft:      "Boeing 777",
			FareClass:     "economy",
			BasePrice:     24900, // $249.00
		},
		{
			ID:            6,
			Origin:        "DFW",
			Destination:   "ORD",
			DepartureTime: now.Add(96 * time.Hour),
			ArrivalTime:   now.Add(96*time.Hour + 2*time.Hour),
			Airline:       "UA",
			Aircraft:      "Boeing 777",
			FareClass:     "first",
			BasePrice:     149900, // $1499.00
		},
	}
}
