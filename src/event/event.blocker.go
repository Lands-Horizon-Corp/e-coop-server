package event

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// Helper to check and update error count/block for an IP
// HandleIPBlocker returns a blocker function and checks if already blocked
func (e *Event) HandleIPBlocker(context context.Context, ctx echo.Context) (blockFn func(reason string), isBlocked bool, err error) {
	realIP := ctx.RealIP()
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
		// Get current count
		count := 0
		countVal, _ := cache.Get(context, errorKey)
		if countVal != nil {
			count, _ = strconv.Atoi(string(countVal))
		}
		count++

		// Update count
		if err := cache.Set(context, errorKey, fmt.Appendf(nil, "%d", count), 5*time.Minute); err != nil {
			return
		}

		// Block if threshold reached
		if count >= 3 {
			if err := cache.Set(context, blockKey, []byte(reason), 5*time.Minute); err != nil {
				return
			}
		}
	}

	return blockFn, false, nil
}
