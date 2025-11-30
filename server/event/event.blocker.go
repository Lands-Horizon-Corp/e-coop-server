package event

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	errorKey := fmt.Sprintf("errorcount:%s", realIP)

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

		// Get current count and timestamp
		count := 0
		var firstErrorTime time.Time
		countVal, _ := cache.Get(context, errorKey)

		if countVal != nil {
			// Parse stored value: "count:timestamp"
			parts := strings.Split(string(countVal), ":")
			if len(parts) == 2 {
				count, _ = strconv.Atoi(parts[0])
				if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					firstErrorTime = time.Unix(timestamp, 0)
				}
			}
		}

		// Reset count if more than 5 minutes have passed since first error
		if !firstErrorTime.IsZero() && now.Sub(firstErrorTime) > 5*time.Minute {
			count = 0
			firstErrorTime = now
		} else if firstErrorTime.IsZero() {
			// First error ever
			firstErrorTime = now
		}

		count++

		// Store count with timestamp: "count:firstErrorTimestamp"
		value := fmt.Sprintf("%d:%d", count, firstErrorTime.Unix())
		if err := cache.Set(context, errorKey, []byte(value), 10*time.Minute); err != nil {
			return
		}

		// Block if threshold reached
		if count >= maxBlockedAttempts {
			if err := cache.Set(context, blockKey, []byte(reason), 5*time.Minute); err != nil {
				return
			}
		}
	}

	return blockFn, false, nil
}
