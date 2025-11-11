package handlers

import (
	"regexp"
	"strings"
)

type RouteHandlerExtractor struct {
	URL string
}

// Constructor
func NewRouteHandlerExtractor(url string) *RouteHandlerExtractor {
	return &RouteHandlerExtractor{URL: url}
}

func (r *RouteHandlerExtractor) MatchableRoute(fn func(params ...string), route string) {
	patternParts := strings.Split(strings.Trim(route, "/"), "/")
	pathParts := strings.Split(strings.Trim(r.URL, "/"), "/")
	regexMatch := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if len(patternParts) != len(pathParts) {
		return
	}
	var params []string
	for i := range patternParts {
		pp := patternParts[i]
		mp := pathParts[i]
		if strings.HasPrefix(pp, ":") {
			if !regexMatch.MatchString(mp) {
				return
			}
			params = append(params, mp)
			continue
		}
		if pp != mp {
			return
		}
	}
	fn(params...)
}
