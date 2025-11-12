package handlers

import (
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

func (r *RouteHandlerExtractor[T]) MatchableRoute(route string, fn func(params ...string) (T, error)) (T, error) {
	pathParts := strings.Split(strings.Trim(r.URL, "/"), "/")
	patternParts := strings.Split(strings.Trim(route, "/"), "/")
	regexMatch := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	var zeroValue T
	if len(patternParts) != len(pathParts) {
		return zeroValue, nil
	}

	var params []string
	for i := range patternParts {
		pp := patternParts[i]
		mp := pathParts[i]
		if strings.HasPrefix(pp, ":") {
			if !regexMatch.MatchString(mp) {
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
