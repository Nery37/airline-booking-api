package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"airline-booking/internal/config"
)

type Router struct {
	engine  *gin.Engine
	handler *BookingHandler
	config  *config.Config
	logger  *zap.Logger
}

func NewRouter(handler *BookingHandler, cfg *config.Config, logger *zap.Logger) *Router {
	// Set gin mode based on environment
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	engine := gin.New()
	
	return &Router{
		engine:  engine,
		handler: handler,
		config:  cfg,
		logger:  logger,
	}
}

func (r *Router) Setup() {
	// Global middleware
	r.engine.Use(r.loggerMiddleware())
	r.engine.Use(r.recoveryMiddleware())
	r.engine.Use(r.corsMiddleware())
	r.engine.Use(r.rateLimitMiddleware())
	
	// API routes
	api := r.engine.Group("/api/v1")
	{
		// Health check endpoint
		api.GET("/health", r.handler.Health)
		
		// Flight search and management
		api.GET("/flights/search", r.handler.SearchFlights)
		api.POST("/flights", r.handler.CreateFlight)
		api.GET("/flights/:flight_id/seats", r.handler.GetFlightSeats)
		
		// Seat holds
		api.POST("/holds", r.handler.CreateHold)
		api.DELETE("/holds/:flight_id/:seat_no", r.handler.ReleaseHold)
		
		// Ticket confirmation
		api.POST("/tickets/confirm", r.handler.ConfirmTicket)
	}
	
	// Debug route without middleware
	r.engine.POST("/debug/holds", r.handler.CreateHold)
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Middleware functions

func (r *Router) loggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			r.logger.Info("HTTP Request",
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.Int("status", param.StatusCode),
				zap.Duration("latency", param.Latency),
				zap.String("ip", param.ClientIP),
			)
			return ""
		},
	})
}

func (r *Router) recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		r.logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(500, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	})
}

func (r *Router) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, User-ID, Idempotency-Key")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func (r *Router) rateLimitMiddleware() gin.HandlerFunc {
	// Create a rate limiter: allow N requests per minute per IP
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(r.config.RateLimit.PerMinute)), r.config.RateLimit.PerMinute)
	
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
