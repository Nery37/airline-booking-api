package models

import (
	"time"
)

// Flight represents a flight in the system
type Flight struct {
	ID            int64     `json:"id" db:"id"`
	Origin        string    `json:"origin" db:"origin"`
	Destination   string    `json:"destination" db:"destination"`
	DepartureTime time.Time `json:"departure_time" db:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time" db:"arrival_time"`
	Airline       string    `json:"airline" db:"airline"`
	Aircraft      string    `json:"aircraft" db:"aircraft"`
	FareClass     string    `json:"fare_class" db:"fare_class"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Seat represents a seat in a flight
type Seat struct {
	ID        int64     `json:"id" db:"id"`
	FlightID  int64     `json:"flight_id" db:"flight_id"`
	SeatNo    string    `json:"seat_no" db:"seat_no"`
	Class     string    `json:"class" db:"class"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SeatLock represents a seat lock/hold
type SeatLock struct {
	ID        int64      `json:"id" db:"id"`
	FlightID  int64      `json:"flight_id" db:"flight_id"`
	SeatNo    string     `json:"seat_no" db:"seat_no"`
	HolderID  string     `json:"holder_id" db:"holder_id"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// Ticket represents a confirmed ticket
type Ticket struct {
	ID          int64     `json:"id" db:"id"`
	FlightID    int64     `json:"flight_id" db:"flight_id"`
	SeatNo      string    `json:"seat_no" db:"seat_no"`
	UserID      string    `json:"user_id" db:"user_id"`
	PriceAmount int64     `json:"price_amount" db:"price_amount"` // in cents
	Currency    string    `json:"currency" db:"currency"`
	IssuedAt    time.Time `json:"issued_at" db:"issued_at"`
	PNRCode     string    `json:"pnr_code" db:"pnr_code"`
	PaymentRef  string    `json:"payment_ref" db:"payment_ref"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// IdempotencyKey represents an idempotency key record
type IdempotencyKey struct {
	RequestID    string    `json:"request_id" db:"request_id"`
	Route        string    `json:"route" db:"route"`
	UserID       string    `json:"user_id" db:"user_id"`
	ResponseHash string    `json:"response_hash" db:"response_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// SeatStatus represents the status of a seat
type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "available"
	SeatStatusHeld      SeatStatus = "held"
	SeatStatusSold      SeatStatus = "sold"
)

// SeatAvailability represents seat availability info
type SeatAvailability struct {
	SeatNo    string     `json:"seat_no"`
	Class     string     `json:"class"`
	Status    SeatStatus `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Price     int64      `json:"price"` // in cents
}

// FlightSearchResult represents a flight search result from Elasticsearch
type FlightSearchResult struct {
	ID            int64               `json:"id"`
	Origin        string              `json:"origin"`
	Destination   string              `json:"destination"`
	DepartureTime time.Time           `json:"departure_time"`
	ArrivalTime   time.Time           `json:"arrival_time"`
	Airline       string              `json:"airline"`
	Aircraft      string              `json:"aircraft"`
	FareClass     string              `json:"fare_class"`
	AvailableSeats int               `json:"available_seats"`
	BasePrice     float64             `json:"base_price"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Hold-related DTOs
type CreateHoldRequest struct {
	FlightID int64  `json:"flight_id" binding:"required"`
	SeatNo   string `json:"seat_no" binding:"required"`
}

type CreateHoldResponse struct {
	FlightID  int64     `json:"flight_id"`
	SeatNo    string    `json:"seat_no"`
	HolderID  string    `json:"holder_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Ticket confirmation DTOs
type ConfirmTicketRequest struct {
	FlightID   int64  `json:"flight_id" binding:"required"`
	SeatNo     string `json:"seat_no" binding:"required"`
	PaymentRef string `json:"payment_ref" binding:"required"`
}

type ConfirmTicketResponse struct {
	TicketID   int64  `json:"ticket_id"`
	FlightID   int64  `json:"flight_id"`
	SeatNo     string `json:"seat_no"`
	PNRCode    string `json:"pnr_code"`
	PaymentRef string `json:"payment_ref"`
}

// Flight search DTOs
type FlightSearchRequest struct {
	Origin      string `form:"origin" binding:"required"`
	Destination string `form:"destination" binding:"required"`
	Date        string `form:"date" binding:"required"` // YYYY-MM-DD format
	FareClass   string `form:"fare_class"`
	Airline     string `form:"airline"`
	Page        int    `form:"page,default=1"`
	Size        int    `form:"size,default=10"`
}

type FlightSearchResponse struct {
	Flights []FlightSearchResult `json:"flights"`
	Total   int64                `json:"total"`
	Page    int                  `json:"page"`
	Size    int                  `json:"size"`
}

// Flight creation DTOs
type CreateFlightRequest struct {
	Origin        string  `json:"origin" binding:"required"`
	Destination   string  `json:"destination" binding:"required"`
	DepartureTime string  `json:"departure_time" binding:"required"` // RFC3339 format
	ArrivalTime   string  `json:"arrival_time" binding:"required"`   // RFC3339 format
	Airline       string  `json:"airline" binding:"required"`
	Aircraft      string  `json:"aircraft" binding:"required"`
	FareClass     string  `json:"fare_class" binding:"required"`
	BasePrice     float64 `json:"base_price" binding:"required"`
	SeatConfig    *SeatConfiguration `json:"seat_config,omitempty"` // Optional seat configuration
}

type SeatConfiguration struct {
	EconomyRows    int `json:"economy_rows" binding:"min=1"`
	BusinessRows   int `json:"business_rows" binding:"min=0"`
	FirstClassRows int `json:"first_class_rows" binding:"min=0"`
	SeatsPerRow    int `json:"seats_per_row" binding:"min=1"`
}

type CreateFlightResponse struct {
	ID            int64              `json:"id"`
	Origin        string             `json:"origin"`
	Destination   string             `json:"destination"`
	DepartureTime string             `json:"departure_time"`
	ArrivalTime   string             `json:"arrival_time"`
	Airline       string             `json:"airline"`
	Aircraft      string             `json:"aircraft"`
	FareClass     string             `json:"fare_class"`
	BasePrice     float64            `json:"base_price"`
	SeatsCreated  int                `json:"seats_created"`
	CreatedAt     string             `json:"created_at"`
}
