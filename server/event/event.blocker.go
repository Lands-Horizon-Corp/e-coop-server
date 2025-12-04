package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

const maxBlockedAttempts = 25

// HandleIPBlocker checks the requestor IP for an existing block and
// returns a block function, a boolean indicating whether the IP is
// already blocked, and an error if the check failed.
func (e *Event) HandleIPBlocker(context context.Context, ctx echo.Context) (blockFn func(reason string), isBlocked bool, err error) {
	realIP := handlers.GetClientIP(ctx)
	cache := e.provider.Service.Cache

	blockKey := fmt.Sprintf("block:%s", realIP)
	errorKey := fmt.Sprintf("transaction_errors:%s", realIP)

	// Check if already blocked
	blockedVal, err := cache.Get(context, blockKey)
	if err != nil {
		return nil, false, eris.Wrap(err, "failed to get block key from cache")
	}
	if blockedVal != nil {
		return nil, true, nil
	}

	// Return the blocking function
	blockFn = func(reason string) {
		now := time.Now().UTC()
		timestamp := float64(now.Unix())

		// Add error attempt to sorted set with timestamp
		if err := cache.ZAdd(context, errorKey, timestamp, fmt.Sprintf("%s:%d", reason, now.Unix())); err != nil {
			return
		}

		// Clean up old errors (keep last 5 minutes for threshold calculation)
		fiveMinutesAgo := now.Add(-5 * time.Minute).Unix()
		if _, err := cache.ZRemRangeByScore(context, errorKey, "0", fmt.Sprintf("%d", fiveMinutesAgo-1)); err != nil {
			// Log but continue
		}

		// Count errors in the last 5 minutes
		count, err := cache.ZCard(context, errorKey)
		if err != nil {
			return
		}

		// Block if threshold reached
		if int(count) >= maxBlockedAttempts {
			// Set block status
			if err := cache.Set(context, blockKey, []byte(reason), 5*time.Minute); err != nil {
				return
			}

			// Track in global transaction blocks registry
			if err := cache.ZAdd(context, "transaction_blocks_registry", timestamp, realIP); err != nil {
				// Log but continue
			}

			// Track block event with details
			blockEventKey := fmt.Sprintf("transaction_block_events:%s", realIP)
			blockData := fmt.Sprintf("%s:%d:%d", reason, count, now.Unix())
			if err := cache.ZAdd(context, blockEventKey, timestamp, blockData); err != nil {
				// Log but continue
			}

			// Clean up old block events (keep last 30 days for analysis)
			thirtyDaysAgo := now.AddDate(0, 0, -30).Unix()
			if _, err := cache.ZRemRangeByScore(context, blockEventKey, "0", fmt.Sprintf("%d", thirtyDaysAgo)); err != nil {
				// Log but continue
			}
		}
	}

	return blockFn, false, nil
}
