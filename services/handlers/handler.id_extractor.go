package handlers

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
)

type RouteHandlerExtractor[T any] struct {
	URL string
}

func NewRouteHandlerExtractor[T any](url string) *RouteHandlerExtractor[T] {
	return &RouteHandlerExtractor[T]{URL: url}
}

func splitPath(s string) []string {
	if s == "" {
		return []string{}
	}
	if idx := strings.IndexAny(s, "?#"); idx != -1 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "/")
}

func (r *RouteHandlerExtractor[T]) MatchableRoute(route string, fn func(params ...string) (T, error)) (T, error) {
	var zeroValue T

	if r == nil {
		return zeroValue, fmt.Errorf("RouteHandlerExtractor receiver is nil")
	}
	if r.URL == "" {
		return zeroValue, fmt.Errorf("extractor URL is empty")
	}
	if route == "" {
		return zeroValue, fmt.Errorf("route is empty")
	}
	if fn == nil {
		return zeroValue, fmt.Errorf("handler function is nil")
	}

	pathParts := splitPath(r.URL)
	patternParts := splitPath(route)

	if len(patternParts) != len(pathParts) {
		return zeroValue, nil
	}
	allowedSegment := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	var params []string
	for i := range patternParts {
		pp := patternParts[i]
		mp := pathParts[i]
		if strings.HasPrefix(pp, ":") {
			if !allowedSegment.MatchString(mp) {
				return zeroValue, nil
			}
			params = append(params, mp)
			continue
		}
		if pp != mp {
			return zeroValue, nil
		}
	}

	var res T
	var err error
	func() {
		defer func() {
			if rcv := recover(); rcv != nil {
				err = fmt.Errorf("panic in route handler: %v\n%s", rcv, debug.Stack())
			}
		}()
		res, err = fn(params...)
	}()
	if err != nil {
		return zeroValue, err
	}
	return res, nil
}
