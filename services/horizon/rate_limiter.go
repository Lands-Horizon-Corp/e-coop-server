// Package horizon provides reusable rate limiting utilities
package horizon

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RateLimiterConfig holds configuration for the rate limiter
type RateLimiterConfig struct {
	// RequestsPerSecond defines the maximum number of requests allowed per second
	RequestsPerSecond int

	// BurstCapacity allows temporary bursts above the normal rate
	BurstCapacity int

	// WindowDuration defines the time window for rate limiting
	WindowDuration time.Duration

	// KeyPrefix is used to namespace rate limit keys in Redis
	KeyPrefix string
}

// DefaultRateLimiterConfig returns a sensible default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 20,
		BurstCapacity:     100,
		WindowDuration:    1 * time.Minute,
		KeyPrefix:         "rate_limit",
	}
}

// RateLimiter provides Redis-backed distributed rate limiting functionality
type RateLimiter struct {
	cache  CacheService
	logger *zap.Logger
	config RateLimiterConfig
}

// NewRateLimiter creates a new RateLimiter instance
func NewRateLimiter(cache CacheService, logger *zap.Logger, config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		cache:  cache,
		logger: logger,
		config: config,
	}
}

// NewRateLimiterWithDefaults creates a new RateLimiter with default configuration
func NewRateLimiterWithDefaults(cache CacheService, logger *zap.Logger) *RateLimiter {
	return NewRateLimiter(cache, logger, DefaultRateLimiterConfig())
}

// Allow checks if a request should be allowed based on the rate limit
// identifier should be a unique string for the entity being rate limited (e.g., IP, user ID, API key)
func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	// Create context with timeout for Redis operations
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowDuration)

	// Generate unique Redis key
	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	// Remove expired entries from the sliding window
	if err := rl.removeExpiredEntries(timeoutCtx, key, windowStart.Unix()); err != nil {
		rl.logger.Error("Failed to clean expired rate limit entries",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err))
		// Continue processing even if cleanup fails
	}

	// Get current request count in the window
	currentCount, err := rl.getRequestCount(timeoutCtx, key)
	if err != nil {
		rl.logger.Error("Rate limit cache error",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err))
		// Allow request on cache errors to prevent service disruption
		return true, nil
	}

	// Calculate maximum requests allowed in the current window
	maxRequests := int(float64(rl.config.RequestsPerSecond) * rl.config.WindowDuration.Seconds())

	if currentCount >= maxRequests {
		rl.logger.Debug("Rate limit exceeded",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Int("current_count", currentCount),
			zap.Int("max_requests", maxRequests),
			zap.Duration("window", rl.config.WindowDuration),
		)
		return false, nil
	}

	// Record the current request
	if err := rl.addRequest(timeoutCtx, key, now.Unix()); err != nil {
		rl.logger.Error("Failed to record rate limit request",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err),
		)
		// Still allow the request even if we can't record it
		return true, nil
	}

	return true, nil
}

// AllowWithDetails returns rate limit status with additional information
func (rl *RateLimiter) AllowWithDetails(ctx context.Context, identifier string) (allowed bool, remaining int, resetTime time.Time, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowDuration)
	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	// Clean expired entries
	if cleanErr := rl.removeExpiredEntries(timeoutCtx, key, windowStart.Unix()); cleanErr != nil {
		rl.logger.Error("Failed to clean expired entries",
			zap.String("identifier", identifier),
			zap.Error(cleanErr))
	}

	// Get current count
	currentCount, err := rl.getRequestCount(timeoutCtx, key)
	if err != nil {
		rl.logger.Error("Rate limit cache error",
			zap.String("identifier", identifier),
			zap.Error(err))
		return true, 0, time.Time{}, err
	}

	maxRequests := int(float64(rl.config.RequestsPerSecond) * rl.config.WindowDuration.Seconds())
	remaining = maxRequests - currentCount
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(rl.config.WindowDuration)
	allowed = currentCount < maxRequests

	if allowed {
		// Record the request
		if addErr := rl.addRequest(timeoutCtx, key, now.Unix()); addErr != nil {
			rl.logger.Error("Failed to record rate limit request",
				zap.String("identifier", identifier),
				zap.Error(addErr))
		} else {
			remaining-- // Decrease remaining count since we recorded this request
		}
	}

	return allowed, remaining, resetTime, nil
}

// GetStatus returns current rate limit status without consuming a request
func (rl *RateLimiter) GetStatus(ctx context.Context, identifier string) (current int, remaining int, resetTime time.Time, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowDuration)
	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	// Clean expired entries
	if cleanErr := rl.removeExpiredEntries(timeoutCtx, key, windowStart.Unix()); cleanErr != nil {
		rl.logger.Debug("Failed to clean expired entries",
			zap.String("identifier", identifier),
			zap.Error(cleanErr))
	}

	current, err = rl.getRequestCount(timeoutCtx, key)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	maxRequests := int(float64(rl.config.RequestsPerSecond) * rl.config.WindowDuration.Seconds())
	remaining = maxRequests - current
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(rl.config.WindowDuration)
	return current, remaining, resetTime, nil
}

// Reset clears the rate limit for a specific identifier
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	if err := rl.cache.Delete(timeoutCtx, key); err != nil {
		rl.logger.Error("Failed to reset rate limit",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	rl.logger.Info("Rate limit reset",
		zap.String("identifier", identifier),
		zap.String("key", key))

	return nil
}

// removeExpiredEntries removes rate limit entries older than the window start time
func (rl *RateLimiter) removeExpiredEntries(ctx context.Context, key string, windowStart int64) error {
	_, err := rl.cache.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart-1))
	if err != nil {
		rl.logger.Debug("Failed to remove expired entries from sorted set",
			zap.String("key", key),
			zap.Int64("window_start", windowStart),
			zap.Error(err))
	}
	return err
}

// getRequestCount returns the number of requests in the current window
func (rl *RateLimiter) getRequestCount(ctx context.Context, key string) (int, error) {
	count, err := rl.cache.ZCard(ctx, key)
	if err != nil {
		rl.logger.Debug("Failed to get request count from sorted set",
			zap.String("key", key),
			zap.Error(err))
		return 0, err
	}
	return int(count), nil
}

// addRequest adds a new request timestamp to the rate limit tracking
func (rl *RateLimiter) addRequest(ctx context.Context, key string, timestamp int64) error {
	err := rl.cache.ZAdd(ctx, key, float64(timestamp), timestamp)
	if err != nil {
		rl.logger.Debug("Failed to add request to sorted set",
			zap.String("key", key),
			zap.Int64("timestamp", timestamp),
			zap.Error(err))
		return err
	}
	return nil
}

// RateLimitMiddleware creates an Echo middleware function using the RateLimiter
func (rl *RateLimiter) RateLimitMiddleware(identifierExtractor func(c echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identifier := identifierExtractor(c)

			allowed, remaining, resetTime, err := rl.AllowWithDetails(c.Request().Context(), identifier)
			if err != nil {
				rl.logger.Error("Rate limiter error",
					zap.String("identifier", identifier),
					zap.Error(err))
				// Allow request on error to prevent service disruption
				return next(c)
			}

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerSecond))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				rl.logger.Warn("Rate limit exceeded",
					zap.String("identifier", identifier),
					zap.Int("remaining", remaining),
					zap.Time("reset_time", resetTime))

				return c.JSON(429, map[string]interface{}{
					"error":       "Too many requests",
					"retry_after": int(time.Until(resetTime).Seconds()),
				})
			}

			return next(c)
		}
	}
}
