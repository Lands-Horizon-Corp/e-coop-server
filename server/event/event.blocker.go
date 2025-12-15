package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

const maxBlockedAttempts = 25

func (e *Event) HandleIPBlocker(
	ctx context.Context,
	echoCtx echo.Context,
) (blockFn func(reason string), isBlocked bool, err error) {

	realIP := handlers.GetClientIP(echoCtx)
	cache := e.provider.Service.Cache
	logger := e.provider.Service.Logger

	blockKey := fmt.Sprintf("block:%s", realIP)
	errorKey := fmt.Sprintf("transaction_errors:%s", realIP)

	logger.Info("ip_blocker_check_start",
		zap.String("ip", realIP),
	)
	blockedVal, err := cache.Get(ctx, blockKey)
	if err != nil {
		logger.Error("ip_blocker_cache_get_failed",
			zap.String("ip", realIP),
			zap.String("key", blockKey),
			zap.Error(err),
		)
		return nil, false, eris.Wrap(err, "failed to get block key from cache")
	}
	if blockedVal != nil {
		logger.Warn("ip_already_blocked",
			zap.String("ip", realIP),
			zap.String("reason", string(blockedVal)),
		)
		return nil, true, nil
	}
	blockFn = func(reason string) {
		now := time.Now().UTC()
		timestamp := float64(now.Unix())

		logger.Warn("ip_blocker_error_recorded",
			zap.String("ip", realIP),
			zap.String("reason", reason),
			zap.Time("time", now),
		)
		if err := cache.ZAdd(
			ctx,
			errorKey,
			timestamp,
			fmt.Sprintf("%s:%d", reason, now.Unix()),
		); err != nil {
			logger.Error("ip_blocker_zadd_failed",
				zap.String("ip", realIP),
				zap.String("key", errorKey),
				zap.Error(err),
			)
			return
		}
		fiveMinutesAgo := now.Add(-5 * time.Minute).Unix()
		if _, err := cache.ZRemRangeByScore(
			ctx,
			errorKey,
			"0",
			fmt.Sprintf("%d", fiveMinutesAgo-1),
		); err != nil {
			logger.Warn("ip_blocker_cleanup_failed",
				zap.String("ip", realIP),
				zap.Error(err),
			)
		}
		count, err := cache.ZCard(ctx, errorKey)
		if err != nil {
			logger.Error("ip_blocker_count_failed",
				zap.String("ip", realIP),
				zap.Error(err),
			)
			return
		}
		logger.Info("ip_blocker_error_count",
			zap.String("ip", realIP),
			zap.Int64("count", count),
			zap.Int("threshold", maxBlockedAttempts),
		)
		if int(count) >= maxBlockedAttempts {
			logger.Warn("ip_blocker_threshold_reached",
				zap.String("ip", realIP),
				zap.Int64("count", count),
				zap.String("reason", reason),
			)
			if err := cache.Set(
				ctx,
				blockKey,
				[]byte(reason),
				5*time.Minute,
			); err != nil {
				logger.Error("ip_blocker_set_failed",
					zap.String("ip", realIP),
					zap.Error(err),
				)
				return
			}
			logger.Warn("ip_blocked",
				zap.String("ip", realIP),
				zap.Duration("duration", 5*time.Minute),
				zap.String("reason", reason),
			)
			if err := cache.ZAdd(
				ctx,
				"transaction_blocks_registry",
				timestamp,
				realIP,
			); err != nil {
				logger.Warn("ip_blocker_registry_add_failed",
					zap.String("ip", realIP),
					zap.Error(err),
				)
			}
			blockEventKey := fmt.Sprintf("transaction_block_events:%s", realIP)
			blockData := fmt.Sprintf("%s:%d:%d", reason, count, now.Unix())
			if err := cache.ZAdd(
				ctx,
				blockEventKey,
				timestamp,
				blockData,
			); err != nil {
				logger.Warn("ip_blocker_event_add_failed",
					zap.String("ip", realIP),
					zap.Error(err),
				)
			}
			thirtyDaysAgo := now.AddDate(0, 0, -30).Unix()
			if _, err := cache.ZRemRangeByScore(
				ctx,
				blockEventKey,
				"0",
				fmt.Sprintf("%d", thirtyDaysAgo),
			); err != nil {
				logger.Warn("ip_blocker_event_cleanup_failed",
					zap.String("ip", realIP),
					zap.Error(err),
				)
			}
		}
	}
	logger.Info("ip_blocker_check_passed",
		zap.String("ip", realIP),
	)
	return blockFn, false, nil
}
