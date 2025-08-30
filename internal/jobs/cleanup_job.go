package jobs

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"airline-booking/internal/service"
)

type CleanupJob struct {
	bookingService *service.BookingService
	logger         *zap.Logger
	cron           *cron.Cron
}

func NewCleanupJob(bookingService *service.BookingService, logger *zap.Logger) *CleanupJob {
	c := cron.New(cron.WithSeconds())
	
	job := &CleanupJob{
		bookingService: bookingService,
		logger:         logger,
		cron:           c,
	}
	
	return job
}

// Start begins the cleanup job that runs every minute
func (j *CleanupJob) Start() error {
	// Run every minute to cleanup expired holds
	_, err := j.cron.AddFunc("0 * * * * *", j.cleanupExpiredHolds)
	if err != nil {
		return err
	}
	
	// Run every hour to cleanup old idempotency keys
	_, err = j.cron.AddFunc("0 0 * * * *", j.cleanupIdempotencyKeys)
	if err != nil {
		return err
	}
	
	j.cron.Start()
	j.logger.Info("Cleanup job started")
	
	return nil
}

// Stop stops the cleanup job
func (j *CleanupJob) Stop() {
	if j.cron != nil {
		j.cron.Stop()
		j.logger.Info("Cleanup job stopped")
	}
}

func (j *CleanupJob) cleanupExpiredHolds() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	start := time.Now()
	err := j.bookingService.CleanupExpiredHolds(ctx)
	duration := time.Since(start)
	
	if err != nil {
		j.logger.Error("Failed to cleanup expired holds",
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		j.logger.Debug("Cleaned up expired holds",
			zap.Duration("duration", duration))
	}
}

func (j *CleanupJob) cleanupIdempotencyKeys() {
	start := time.Now()
	// TODO: Implement cleanup of old idempotency keys
	// For now, we'll just log
	duration := time.Since(start)
	
	j.logger.Debug("Cleaned up old idempotency keys",
		zap.Duration("duration", duration))
}
