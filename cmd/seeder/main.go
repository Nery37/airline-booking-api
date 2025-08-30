package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"airline-booking/internal/config"
	"airline-booking/internal/db"
	"airline-booking/internal/es"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to database
	database, err := db.NewDatabase(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	// Connect to Elasticsearch
	esClient, err := es.NewClient(&cfg.Elasticsearch, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Elasticsearch", zap.Error(err))
	}

	logger.Info("Starting database seeding...")

	// Seed database
	if err := seedDatabase(database); err != nil {
		logger.Fatal("Failed to seed database", zap.Error(err))
	}

	logger.Info("Database seeded successfully")

	// Seed Elasticsearch
	logger.Info("Starting Elasticsearch seeding...")
	if err := seedElasticsearch(esClient, database, logger); err != nil {
		logger.Fatal("Failed to seed Elasticsearch", zap.Error(err))
	}

	logger.Info("Elasticsearch seeded successfully")
	logger.Info("Seeding completed successfully!")
}

func seedDatabase(database *db.Database) error {
	// Check if data already exists
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM flights").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		fmt.Printf("Database already has %d flights, skipping database seed\n", count)
		return nil
	}

	// Read and execute seed file
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute seed data - we'll execute it in chunks to avoid issues
	seedQueries := []string{
		// Insert flights
		`INSERT INTO flights (id, origin, destination, departure_time, arrival_time, airline, aircraft, fare_class) VALUES
		(1, 'JFK', 'LAX', '2025-08-30 10:30:00', '2025-08-30 14:30:00', 'AA', 'Boeing 737', 'economy'),
		(2, 'JFK', 'LAX', '2025-08-30 15:45:00', '2025-08-30 19:45:00', 'DL', 'Airbus A320', 'economy'),
		(3, 'LAX', 'JFK', '2025-08-30 08:15:00', '2025-08-30 16:45:00', 'AA', 'Boeing 777', 'business'),
		(4, 'LAX', 'JFK', '2025-08-30 22:00:00', '2025-08-31 06:30:00', 'UA', 'Boeing 787', 'economy'),
		(5, 'MIA', 'LAS', '2025-08-30 11:20:00', '2025-08-30 14:10:00', 'SW', 'Boeing 737', 'economy'),
		(6, 'LAS', 'MIA', '2025-08-30 16:30:00', '2025-08-30 23:45:00', 'SW', 'Boeing 737', 'economy'),
		(7, 'ORD', 'SFO', '2025-08-30 09:00:00', '2025-08-30 11:30:00', 'UA', 'Boeing 757', 'economy'),
		(8, 'SFO', 'ORD', '2025-08-30 13:15:00', '2025-08-30 19:20:00', 'AA', 'Airbus A321', 'business'),
		(9, 'ATL', 'SEA', '2025-08-30 07:45:00', '2025-08-30 10:15:00', 'DL', 'Boeing 767', 'economy'),
		(10, 'SEA', 'ATL', '2025-08-30 18:30:00', '2025-08-31 02:10:00', 'DL', 'Boeing 767', 'first'),
		(11, 'JFK', 'LHR', '2025-08-30 21:00:00', '2025-08-31 08:30:00', 'BA', 'Boeing 777', 'business'),
		(12, 'LHR', 'JFK', '2025-08-30 14:20:00', '2025-08-30 17:45:00', 'VS', 'Airbus A350', 'economy'),
		(13, 'LAX', 'NRT', '2025-08-30 11:45:00', '2025-08-31 16:20:00', 'JL', 'Boeing 787', 'business'),
		(14, 'NRT', 'LAX', '2025-08-30 17:30:00', '2025-08-30 10:15:00', 'ANA', 'Boeing 777', 'economy'),
		(15, 'MIA', 'GRU', '2025-08-30 23:55:00', '2025-08-31 09:40:00', 'AA', 'Boeing 777', 'economy')`,
	}

	// Execute seed queries
	for _, query := range seedQueries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute seed query: %w", err)
		}
	}

	// Insert seats for main flights
	seatQueries := generateSeatQueries()
	for _, query := range seatQueries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to insert seats: %w", err)
		}
	}

	// Insert sample holds and tickets
	holdsAndTicketsQueries := []string{
		`INSERT INTO seat_locks (flight_id, seat_no, holder_id, expires_at) VALUES
		(1, '12A', 'user123', DATE_ADD(NOW(), INTERVAL 10 MINUTE)),
		(1, '15C', 'user456', DATE_ADD(NOW(), INTERVAL 5 MINUTE)),
		(2, '8B', 'user789', DATE_ADD(NOW(), INTERVAL 12 MINUTE)),
		(3, '2A', 'user111', DATE_ADD(NOW(), INTERVAL 8 MINUTE))`,

		`INSERT INTO tickets (flight_id, seat_no, user_id, price_amount, currency, pnr_code, payment_ref) VALUES
		(1, '10A', 'customer001', 29900, 'USD', 'ABC001', 'pay_001_12345'),
		(1, '10B', 'customer002', 29900, 'USD', 'ABC002', 'pay_002_12346'),
		(1, '1A', 'customer003', 149900, 'USD', 'ABC003', 'pay_003_12347'),
		(2, '5A', 'customer004', 31900, 'USD', 'DEF001', 'pay_004_12348'),
		(2, '5B', 'customer005', 31900, 'USD', 'DEF002', 'pay_005_12349'),
		(3, '1A', 'customer006', 89900, 'USD', 'GHI001', 'pay_006_12350'),
		(3, '10A', 'customer007', 49900, 'USD', 'GHI002', 'pay_007_12351'),
		(4, '15A', 'customer008', 27900, 'USD', 'JKL001', 'pay_008_12352'),
		(5, '5A', 'customer009', 19900, 'USD', 'MNO001', 'pay_009_12353')`,
	}

	for _, query := range holdsAndTicketsQueries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to insert holds/tickets: %w", err)
		}
	}

	return tx.Commit()
}

func generateSeatQueries() []string {
	var queries []string

	// Flight 1: Boeing 737 - detailed seat map
	flight1Seats := []string{}
	// First class (rows 1-3)
	for row := 1; row <= 3; row++ {
		for _, seat := range []string{"A", "B", "C", "D"} {
			flight1Seats = append(flight1Seats, fmt.Sprintf("(1, '%d%s', 'first')", row, seat))
		}
	}
	// Business class (rows 4-8)
	for row := 4; row <= 8; row++ {
		for _, seat := range []string{"A", "B", "C", "D"} {
			flight1Seats = append(flight1Seats, fmt.Sprintf("(1, '%d%s', 'business')", row, seat))
		}
	}
	// Economy class (rows 9-30)
	for row := 9; row <= 30; row++ {
		for _, seat := range []string{"A", "B", "C", "D", "E", "F"} {
			flight1Seats = append(flight1Seats, fmt.Sprintf("(1, '%d%s', 'economy')", row, seat))
		}
	}

	if len(flight1Seats) > 0 {
		queries = append(queries, "INSERT INTO seats (flight_id, seat_no, class) VALUES "+strings.Join(flight1Seats, ", "))
	}

	// Flight 2: Airbus A320 - simplified
	flight2Seats := []string{}
	// Business class (rows 1-4)
	for row := 1; row <= 4; row++ {
		for _, seat := range []string{"A", "B", "C", "D"} {
			flight2Seats = append(flight2Seats, fmt.Sprintf("(2, '%d%s', 'business')", row, seat))
		}
	}
	// Economy class (rows 5-25)
	for row := 5; row <= 25; row++ {
		for _, seat := range []string{"A", "B", "C", "D", "E", "F"} {
			flight2Seats = append(flight2Seats, fmt.Sprintf("(2, '%d%s', 'economy')", row, seat))
		}
	}

	if len(flight2Seats) > 0 {
		queries = append(queries, "INSERT INTO seats (flight_id, seat_no, class) VALUES "+strings.Join(flight2Seats, ", "))
	}

	// Add some seats for other flights (simplified)
	otherFlightSeats := []string{
		"INSERT INTO seats (flight_id, seat_no, class) VALUES (3, '1A', 'business'), (3, '1B', 'business'), (3, '1C', 'business'), (3, '1D', 'business'), (3, '2A', 'business'), (3, '2B', 'business'), (3, '10A', 'economy'), (3, '10B', 'economy'), (3, '10C', 'economy'), (3, '11A', 'economy'), (3, '11B', 'economy'), (3, '12A', 'economy')",
		"INSERT INTO seats (flight_id, seat_no, class) VALUES (4, '15A', 'economy'), (4, '15B', 'economy'), (4, '15C', 'economy'), (4, '16A', 'economy'), (4, '16B', 'economy'), (4, '17A', 'economy')",
		"INSERT INTO seats (flight_id, seat_no, class) VALUES (5, '5A', 'economy'), (5, '5B', 'economy'), (5, '5C', 'economy'), (5, '6A', 'economy'), (5, '6B', 'economy'), (5, '7A', 'economy')",
	}

	queries = append(queries, otherFlightSeats...)

	return queries
}

func seedElasticsearch(esClient *es.Client, database *db.Database, logger *zap.Logger) error {
	// Create all indices
	if err := esClient.CreateIndex(context.Background()); err != nil {
		return fmt.Errorf("failed to create elasticsearch indices: %w", err)
	}

	// Check if flights already exist
	flightCount, err := esClient.CountDocuments("flights")
	if err != nil {
		logger.Warn("Failed to count flights", zap.Error(err))
	}

	if flightCount > 0 {
		logger.Info("Elasticsearch already has documents, skipping ES seed", zap.Int64("flight_count", flightCount))
		return nil
	}

	// Seed flights
	if err := seedFlightsToES(esClient, database, logger); err != nil {
		return fmt.Errorf("failed to seed flights: %w", err)
	}

	// Seed holds
	if err := seedHoldsToES(esClient, database, logger); err != nil {
		return fmt.Errorf("failed to seed holds: %w", err)
	}

	// Seed tickets
	if err := seedTicketsToES(esClient, database, logger); err != nil {
		return fmt.Errorf("failed to seed tickets: %w", err)
	}

	logger.Info("Elasticsearch seeding completed successfully")
	return nil
}

func seedFlightsToES(esClient *es.Client, database *db.Database, logger *zap.Logger) error {
	// Fetch flights from database
	flights, err := getFlightsFromDB(database)
	if err != nil {
		return fmt.Errorf("failed to get flights from database: %w", err)
	}

	// Convert to FlightDocument and index
	flightDocs := make([]es.FlightDocument, len(flights))
	for i, flight := range flights {
		flightDocs[i] = es.FlightDocument{
			ID:            flight["id"].(int64),
			Origin:        flight["origin"].(string),
			Destination:   flight["destination"].(string),
			DepartureTime: flight["departure_time"].(time.Time),
			ArrivalTime:   flight["arrival_time"].(time.Time),
			Airline:       flight["airline"].(string),
			Aircraft:      flight["aircraft"].(string),
			FareClass:     flight["fare_class"].(string),
			BasePrice:     399.99, // Default price
		}
	}

	if err := esClient.BulkIndexFlights(context.Background(), flightDocs); err != nil {
		return fmt.Errorf("failed to bulk index flights: %w", err)
	}

	logger.Info("Indexed flights in Elasticsearch", zap.Int("count", len(flightDocs)))
	return nil
}

func seedHoldsToES(esClient *es.Client, database *db.Database, logger *zap.Logger) error {
	holds, err := getHoldsFromDB(database)
	if err != nil {
		return fmt.Errorf("failed to get holds from database: %w", err)
	}

	// Index each hold individually
	for _, hold := range holds {
		holdDoc := es.HoldDocument{
			ID:        hold["id"].(int64),
			FlightID:  hold["flight_id"].(int64),
			SeatNo:    hold["seat_no"].(string),
			HolderID:  hold["holder_id"].(string),
			ExpiresAt: hold["expires_at"].(*time.Time),
			CreatedAt: hold["created_at"].(time.Time),
			UpdatedAt: hold["updated_at"].(time.Time),
			Status:    "active",
		}

		if err := esClient.IndexHold(context.Background(), holdDoc); err != nil {
			logger.Warn("Failed to index hold", zap.Error(err), zap.Int64("hold_id", holdDoc.ID))
		}
	}

	logger.Info("Indexed holds in Elasticsearch", zap.Int("count", len(holds)))
	return nil
}

func seedTicketsToES(esClient *es.Client, database *db.Database, logger *zap.Logger) error {
	tickets, err := getTicketsFromDB(database)
	if err != nil {
		return fmt.Errorf("failed to get tickets from database: %w", err)
	}

	// Index each ticket individually
	for _, ticket := range tickets {
		ticketDoc := es.TicketDocument{
			ID:          ticket["id"].(int64),
			FlightID:    ticket["flight_id"].(int64),
			SeatNo:      ticket["seat_no"].(string),
			UserID:      ticket["user_id"].(string),
			PriceAmount: ticket["price_amount"].(int64),
			Currency:    ticket["currency"].(string),
			IssuedAt:    ticket["issued_at"].(time.Time),
			PnrCode:     ticket["pnr_code"].(string),
			PaymentRef:  ticket["payment_ref"].(string),
			CreatedAt:   ticket["created_at"].(time.Time),
			Status:      "confirmed",
		}

		if err := esClient.IndexTicket(context.Background(), ticketDoc); err != nil {
			logger.Warn("Failed to index ticket", zap.Error(err), zap.Int64("ticket_id", ticketDoc.ID))
		}
	}

	logger.Info("Indexed tickets in Elasticsearch", zap.Int("count", len(tickets)))
	return nil
}

func getFlightsFromDB(database *db.Database) ([]map[string]interface{}, error) {
	query := `
		SELECT id, origin, destination, departure_time, arrival_time, 
		       airline, aircraft, fare_class, created_at, updated_at
		FROM flights
		ORDER BY id
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []map[string]interface{}
	for rows.Next() {
		var id int64
		var origin, destination, airline, aircraft, fareClass string
		var departureTime, arrivalTime, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &origin, &destination, &departureTime, &arrivalTime,
			&airline, &aircraft, &fareClass, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		flight := map[string]interface{}{
			"id":             id,
			"origin":         origin,
			"destination":    destination,
			"departure_time": departureTime,
			"arrival_time":   arrivalTime,
			"airline":        airline,
			"aircraft":       aircraft,
			"fare_class":     fareClass,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}

		flights = append(flights, flight)
	}

	return flights, rows.Err()
}

func getHoldsFromDB(database *db.Database) ([]map[string]interface{}, error) {
	query := `
		SELECT id, flight_id, seat_no, holder_id, expires_at, created_at, updated_at
		FROM seat_locks
		WHERE expires_at IS NOT NULL AND expires_at > NOW()
		ORDER BY id
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holds []map[string]interface{}
	for rows.Next() {
		var id, flightID int64
		var seatNo, holderID string
		var expiresAt *time.Time
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &flightID, &seatNo, &holderID, &expiresAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		hold := map[string]interface{}{
			"id":         id,
			"flight_id":  flightID,
			"seat_no":    seatNo,
			"holder_id":  holderID,
			"expires_at": expiresAt,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}

		holds = append(holds, hold)
	}

	return holds, rows.Err()
}

func getTicketsFromDB(database *db.Database) ([]map[string]interface{}, error) {
	query := `
		SELECT id, flight_id, seat_no, user_id, price_amount, currency, 
		       issued_at, pnr_code, payment_ref, created_at
		FROM tickets
		ORDER BY id
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []map[string]interface{}
	for rows.Next() {
		var id, flightID, priceAmount int64
		var seatNo, userID, currency, pnrCode, paymentRef string
		var issuedAt, createdAt time.Time

		err := rows.Scan(&id, &flightID, &seatNo, &userID, &priceAmount, &currency,
			&issuedAt, &pnrCode, &paymentRef, &createdAt)
		if err != nil {
			return nil, err
		}

		ticket := map[string]interface{}{
			"id":           id,
			"flight_id":    flightID,
			"seat_no":      seatNo,
			"user_id":      userID,
			"price_amount": priceAmount,
			"currency":     currency,
			"issued_at":    issuedAt,
			"pnr_code":     pnrCode,
			"payment_ref":  paymentRef,
			"created_at":   createdAt,
		}

		tickets = append(tickets, ticket)
	}

	return tickets, rows.Err()
}
