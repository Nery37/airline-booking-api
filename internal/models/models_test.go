package models

import (
	"testing"
	"time"
)

func TestFlightValidation(t *testing.T) {
	flight := Flight{
		Origin:        "JFK",
		Destination:   "LAX", 
		DepartureTime: time.Now().Add(24 * time.Hour),
		ArrivalTime:   time.Now().Add(29 * time.Hour),
		Airline:       "AA",
		Aircraft:      "Boeing 737",
		FareClass:     "economy",
	}

	if flight.Origin == "" {
		t.Error("Origin should not be empty")
	}

	if flight.Destination == "" {
		t.Error("Destination should not be empty")
	}

	if flight.DepartureTime.After(flight.ArrivalTime) {
		t.Error("Departure time should be before arrival time")
	}
}

func TestCreateHoldRequest(t *testing.T) {
	req := CreateHoldRequest{
		FlightID: 1,
		SeatNo:   "12A",
	}

	if req.FlightID <= 0 {
		t.Error("FlightID should be positive")
	}

	if req.SeatNo == "" {
		t.Error("SeatNo should not be empty")
	}
}

func TestConfirmTicketRequest(t *testing.T) {
	req := ConfirmTicketRequest{
		FlightID:   1,
		SeatNo:     "12A",
		PaymentRef: "pay_123",
	}

	if req.FlightID <= 0 {
		t.Error("FlightID should be positive")
	}

	if req.SeatNo == "" {
		t.Error("SeatNo should not be empty")
	}

	if req.PaymentRef == "" {
		t.Error("PaymentRef should not be empty")
	}
}
