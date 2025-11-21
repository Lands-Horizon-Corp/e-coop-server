package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

type RouteHandlerExtractor[T any] struct {
	URL string
}

// Constructor
func NewRouteHandlerExtractor[T any](url string) *RouteHandlerExtractor[T] {
	return &RouteHandlerExtractor[T]{URL: url}
}

var allowedSegment = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// splitPath normalizes a URL/route by removing query/fragment, trimming slashes
// and returning path segments. It handles values with or without a leading slash,
// and the root path "/" becomes an empty slice.
func splitPath(s string) []string {
	if s == "" {
		return []string{}
	}
	// remove query and fragment if present
	if idx := strings.IndexAny(s, "?#"); idx != -1 {
		s = s[:idx]
	}
	// trim whitespace and slashes
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "/")
}

func (r *RouteHandlerExtractor[T]) MatchableRoute(route string, fn func(params ...string) (T, error)) (T, error) {
	var zeroValue T

	// defensive checks to avoid nil deref panics
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
		// not a match (preserve previous behavior of returning zero value and no error)
		return zeroValue, nil
	}

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
	return fn(params...)
}
