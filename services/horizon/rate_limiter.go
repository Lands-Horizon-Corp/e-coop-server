package horizon

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type RateLimiterConfig struct {
	RequestsPerSecond int

	RequestsPerMinute int

	BurstCapacity int

	WindowDuration time.Duration

	KeyPrefix string
}

func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 0,  // Use requests per minute instead
		RequestsPerMinute: 60, // 60 requests per minute
		BurstCapacity:     100,
		WindowDuration:    1 * time.Minute,
		KeyPrefix:         "rate_limit",
	}
}

type RateLimiter struct {
	cache  CacheService
	logger *zap.Logger
	config RateLimiterConfig
}

func (rl *RateLimiter) getEffectiveRequestRate() float64 {
	if rl.config.RequestsPerMinute > 0 {
		return float64(rl.config.RequestsPerMinute) / 60.0
	}
	return float64(rl.config.RequestsPerSecond)
}

func NewRateLimiter(cache CacheService, logger *zap.Logger, config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		cache:  cache,
		logger: logger,
		config: config,
	}
}

func NewRateLimiterWithDefaults(cache CacheService, logger *zap.Logger) *RateLimiter {
	return NewRateLimiter(cache, logger, DefaultRateLimiterConfig())
}

func NewRateLimiterPerMinute(cache CacheService, logger *zap.Logger, requestsPerMinute int) *RateLimiter {
	config := RateLimiterConfig{
		RequestsPerSecond: 0,
		RequestsPerMinute: requestsPerMinute,
		BurstCapacity:     requestsPerMinute * 2, // Allow burst of 2x the per-minute rate
		WindowDuration:    1 * time.Minute,
		KeyPrefix:         "rate_limit",
	}
	return NewRateLimiter(cache, logger, config)
}

func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowDuration)

	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	if err := rl.removeExpiredEntries(timeoutCtx, key, windowStart.Unix()); err != nil {
		rl.logger.Error("Failed to clean expired rate limit entries",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err))
	}

	currentCount, err := rl.getRequestCount(timeoutCtx, key)
	if err != nil {
		rl.logger.Error("Rate limit cache error",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err))
		return true, nil
	}

	effectiveRate := rl.getEffectiveRequestRate()
	maxRequests := int(effectiveRate * rl.config.WindowDuration.Seconds())

	if currentCount >= maxRequests {
		rl.logger.Debug("Rate limit exceeded",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Int("current_count", currentCount),
			zap.Int("max_requests", maxRequests),
			zap.Float64("effective_rate", effectiveRate),
			zap.Int("requests_per_minute", rl.config.RequestsPerMinute),
			zap.Duration("window", rl.config.WindowDuration),
		)
		return false, nil
	}

	if err := rl.addRequest(timeoutCtx, key, now.Unix()); err != nil {
		rl.logger.Error("Failed to record rate limit request",
			zap.String("identifier", identifier),
			zap.String("key", key),
			zap.Error(err),
		)
		return true, nil
	}

	return true, nil
}

func (rl *RateLimiter) AllowWithDetails(ctx context.Context, identifier string) (allowed bool, remaining int, resetTime time.Time, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowDuration)
	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	if cleanErr := rl.removeExpiredEntries(timeoutCtx, key, windowStart.Unix()); cleanErr != nil {
		rl.logger.Error("Failed to clean expired entries",
			zap.String("identifier", identifier),
			zap.Error(cleanErr))
	}

	currentCount, err := rl.getRequestCount(timeoutCtx, key)
	if err != nil {
		rl.logger.Error("Rate limit cache error",
			zap.String("identifier", identifier),
			zap.Error(err))
		return true, 0, time.Time{}, err
	}

	maxRequests := int(rl.getEffectiveRequestRate() * rl.config.WindowDuration.Seconds())
	remaining = maxRequests - currentCount
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(rl.config.WindowDuration)
	allowed = currentCount < maxRequests

	if allowed {
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

func (rl *RateLimiter) GetStatus(ctx context.Context, identifier string) (current int, remaining int, resetTime time.Time, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowDuration)
	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

	if cleanErr := rl.removeExpiredEntries(timeoutCtx, key, windowStart.Unix()); cleanErr != nil {
		rl.logger.Debug("Failed to clean expired entries",
			zap.String("identifier", identifier),
			zap.Error(cleanErr))
	}

	current, err = rl.getRequestCount(timeoutCtx, key)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	maxRequests := int(rl.getEffectiveRequestRate() * rl.config.WindowDuration.Seconds())
	remaining = maxRequests - current
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(rl.config.WindowDuration)
	return current, remaining, resetTime, nil
}

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

func (rl *RateLimiter) RateLimitMiddleware(identifierExtractor func(c echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identifier := identifierExtractor(c)

			allowed, remaining, resetTime, err := rl.AllowWithDetails(c.Request().Context(), identifier)
			if err != nil {
				rl.logger.Error("Rate limiter error",
					zap.String("identifier", identifier),
					zap.Error(err))
				return next(c)
			}

			if rl.config.RequestsPerMinute > 0 {
				c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d per minute", rl.config.RequestsPerMinute))
			} else {
				c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d per second", rl.config.RequestsPerSecond))
			}
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
