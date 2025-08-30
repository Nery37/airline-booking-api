package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"airline-booking/internal/models"
	"airline-booking/internal/service"
)

type BookingHandler struct {
	bookingService *service.BookingService
	logger         *zap.Logger
}

func NewBookingHandler(bookingService *service.BookingService, logger *zap.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		logger:         logger,
	}
}

// CreateHold godoc
// @Summary Create a seat hold
// @Description Create a hold on a specific seat for 15 minutes
// @Tags holds
// @Accept json
// @Produce json
// @Param Idempotency-Key header string false "Idempotency key for request deduplication"
// @Param User-ID header string true "User ID for the hold"
// @Param request body models.CreateHoldRequest true "Hold request"
// @Success 201 {object} models.CreateHoldResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /holds [post]
func (h *BookingHandler) CreateHold(c *gin.Context) {
	var req models.CreateHoldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}
	
	userID := c.GetHeader("User-ID")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User-ID header is required", nil)
		return
	}
	
	idempotencyKey := c.GetHeader("Idempotency-Key")
	
	h.logger.Info("CreateHold request received", 
		zap.Int64("flight_id", req.FlightID), 
		zap.String("seat_no", req.SeatNo),
		zap.String("user_id", userID))
	
	response, err := h.bookingService.CreateHold(c.Request.Context(), req, userID, idempotencyKey)
	if err != nil {
		if err.Error() == "seat is already held by another user" || err.Error() == "seat is already sold" {
			h.respondError(c, http.StatusConflict, "SEAT_UNAVAILABLE", err.Error(), nil)
			return
		}
		h.logger.Error("Failed to create hold", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create hold", nil)
		return
	}
	
	c.JSON(http.StatusCreated, response)
}

// ReleaseHold godoc
// @Summary Release a seat hold
// @Description Release a hold on a specific seat
// @Tags holds
// @Param User-ID header string true "User ID who owns the hold"
// @Param flight_id path int true "Flight ID"
// @Param seat_no path string true "Seat number"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /holds/{flight_id}/{seat_no} [delete]
func (h *BookingHandler) ReleaseHold(c *gin.Context) {
	userID := c.GetHeader("User-ID")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User-ID header is required", nil)
		return
	}
	
	flightIDStr := c.Param("flight_id")
	flightID, err := strconv.ParseInt(flightIDStr, 10, 64)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_FLIGHT_ID", "Invalid flight ID", nil)
		return
	}
	
	seatNo := c.Param("seat_no")
	if seatNo == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_SEAT_NO", "Seat number is required", nil)
		return
	}
	
	err = h.bookingService.ReleaseHold(c.Request.Context(), flightID, seatNo, userID)
	if err != nil {
		h.logger.Error("Failed to release hold", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to release hold", nil)
		return
	}
	
	c.Status(http.StatusNoContent)
}

// ConfirmTicket godoc
// @Summary Confirm a ticket purchase
// @Description Confirm a held seat and create a ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Param Idempotency-Key header string false "Idempotency key for request deduplication"
// @Param User-ID header string true "User ID for the ticket"
// @Param request body models.ConfirmTicketRequest true "Ticket confirmation request"
// @Success 201 {object} models.ConfirmTicketResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tickets/confirm [post]
func (h *BookingHandler) ConfirmTicket(c *gin.Context) {
	var req models.ConfirmTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}
	
	userID := c.GetHeader("User-ID")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "User-ID header is required", nil)
		return
	}
	
	idempotencyKey := c.GetHeader("Idempotency-Key")
	
	response, err := h.bookingService.ConfirmTicket(c.Request.Context(), req, userID, idempotencyKey)
	if err != nil {
		if err.Error() == "no valid hold found to confirm" {
			h.respondError(c, http.StatusConflict, "NO_VALID_HOLD", err.Error(), nil)
			return
		}
		h.logger.Error("Failed to confirm ticket", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to confirm ticket", nil)
		return
	}
	
	c.JSON(http.StatusCreated, response)
}

// GetFlightSeats godoc
// @Summary Get flight seat availability
// @Description Get the availability status of all seats for a flight
// @Tags flights
// @Param flight_id path int true "Flight ID"
// @Success 200 {array} models.SeatAvailability
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /flights/{flight_id}/seats [get]
func (h *BookingHandler) GetFlightSeats(c *gin.Context) {
	flightIDStr := c.Param("flight_id")
	flightID, err := strconv.ParseInt(flightIDStr, 10, 64)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_FLIGHT_ID", "Invalid flight ID", nil)
		return
	}
	
	seats, err := h.bookingService.GetFlightSeatAvailability(c.Request.Context(), flightID)
	if err != nil {
		h.logger.Error("Failed to get flight seats", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get flight seats", nil)
		return
	}
	
	c.JSON(http.StatusOK, seats)
}

// SearchFlights godoc
// @Summary Search for flights
// @Description Search for flights using various criteria
// @Tags flights
// @Param origin query string true "Origin airport code"
// @Param destination query string true "Destination airport code"
// @Param date query string true "Departure date (YYYY-MM-DD)"
// @Param fare_class query string false "Fare class"
// @Param airline query string false "Airline code"
// @Param page query int false "Page number (default: 1)"
// @Param size query int false "Page size (default: 10)"
// @Success 200 {object} models.FlightSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /flights/search [get]
func (h *BookingHandler) SearchFlights(c *gin.Context) {
	var req models.FlightSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid query parameters", err.Error())
		return
	}
	
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	if req.Size > 100 {
		req.Size = 100 // Limit page size
	}
	
	response, err := h.bookingService.SearchFlights(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search flights", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search flights", nil)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// CreateFlight godoc
// @Summary Create a new flight
// @Description Create a new flight and automatically index it in Elasticsearch
// @Tags flights
// @Accept json
// @Produce json
// @Param request body models.CreateFlightRequest true "Flight creation request"
// @Success 201 {object} models.CreateFlightResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /flights [post]
func (h *BookingHandler) CreateFlight(c *gin.Context) {
	var req models.CreateFlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}
	
	response, err := h.bookingService.CreateFlight(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create flight", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create flight", nil)
		return
	}
	
	c.JSON(http.StatusCreated, response)
}

// Health godoc
// @Summary Health check
// @Description Check the health of the service
// @Tags health
// @Success 200 {object} map[string]string
// @Router /api/v1/health [get]
func (h *BookingHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   c.Request.Header.Get("X-Request-Time"),
	})
}

func (h *BookingHandler) respondError(c *gin.Context, statusCode int, code, message string, details interface{}) {
	response := models.ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
	c.JSON(statusCode, response)
}
