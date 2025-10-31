package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/rotisserie/eris"
)

// Route defines an API endpoint with its metadata
type Route struct {
	Route        string `json:"route"`
	Request      string `json:"request,omitempty"`  // TypeScript interface for request
	Response     string `json:"response,omitempty"` // TypeScript interface for response
	RequestType  any    // Go type for request (used internally)
	ResponseType any    // Go type for response (used internally)
	Method       string `json:"method"`            // HTTP method (GET, POST, etc.)
	Note         string `json:"note"`              // Additional documentation
	Private      bool   `json:"private,omitempty"` // Excluded from public docs
}

// GroupedRoute organizes routes by their base path segment
type GroupedRoute struct {
	Key    string  `json:"key"`    // Base path segment (e.g. "users")
	Routes []Route `json:"routes"` // Routes under this group
}

// APIInterfaces represents a TypeScript interface definition
type APIInterfaces struct {
	Key   string `json:"key"`   // Interface name
	Value string `json:"value"` // Full TypeScript definition
}

// API aggregates all documentation components
type API struct {
	GroupedRoutes []GroupedRoute  `json:"grouped_routes"` // Routes grouped by base path
	Requests      []APIInterfaces `json:"requests"`       // Unique request interfaces
	Responses     []APIInterfaces `json:"responses"`      // Unique response interfaces
}

const (
	exportInterfacePrefix = "export interface "
	interfacePrefix       = "interface "
)

// extractInterfaceNameFromTS parses TypeScript to extract interface name
func extractInterfaceNameFromTS(tsDefinition string) string {
	lines := strings.SplitSeq(tsDefinition, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, exportInterfacePrefix):
			return extractName(trimmed, exportInterfacePrefix)
		case strings.HasPrefix(trimmed, interfacePrefix):
			return extractName(trimmed, interfacePrefix)
		}
	}
	return ""
}

// extractName helper for interface name extraction
func extractName(line, prefix string) string {
	afterPrefix := strings.TrimPrefix(line, prefix)
	if name := strings.Fields(afterPrefix); len(name) > 0 {
		return name[0]
	}
	return ""
}

// getUniqueInterfaces filters unique interfaces from routes
func getUniqueInterfaces(routes []Route, fieldExtractor func(*Route) string) []APIInterfaces {
	seen := make(map[string]struct{})
	var unique []APIInterfaces

	for _, route := range routes {
		tsDefinition := fieldExtractor(&route)
		if tsDefinition == "" || tsDefinition == "None" {
			continue
		}

		name := extractInterfaceNameFromTS(tsDefinition)
		if name == "" {
			continue
		}

		// Skip duplicates
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}

		unique = append(unique, APIInterfaces{
			Key:   name,
			Value: tsDefinition,
		})
	}
	return unique
}

// GetAllRequestInterfaces gets unique request interfaces
func GetAllRequestInterfaces(routes []Route) []APIInterfaces {
	extractor := func(rt *Route) string { return rt.Request }
	return getUniqueInterfaces(routes, extractor)
}

// GetAllResponseInterfaces gets unique response interfaces
func GetAllResponseInterfaces(routes []Route) []APIInterfaces {
	extractor := func(rt *Route) string { return rt.Response }
	return getUniqueInterfaces(routes, extractor)
}

// RouteHandler manages route registration and documentation
type RouteHandler struct {
	RoutesList     []Route
	interfacesList []APIInterfaces
}

// NewRouteHandler creates a new route manager
func NewRouteHandler() *RouteHandler {
	return &RouteHandler{
		RoutesList:     []Route{},
		interfacesList: []APIInterfaces{},
	}
}

// AddRoute registers a new API endpoint
func (h *RouteHandler) AddRoute(route Route) error {
	method := strings.ToUpper(strings.TrimSpace(route.Method))

	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		// OK
		break
	default:
		return eris.Errorf("unsupported HTTP method: %s for route: %s", method, route.Route)
	}
	for _, existing := range h.RoutesList {
		if strings.EqualFold(existing.Route, route.Route) &&
			strings.EqualFold(existing.Method, method) {
			return eris.Errorf("route already registered: %s %s", method, route.Route)
		}
	}
	// Skip private routes
	if route.Private {
		return nil
	}

	// Convert Go types to TypeScript definitions
	tsRequest := TagFormat(route.RequestType)
	tsResponse := TagFormat(route.ResponseType)

	// Store interfaces
	h.interfacesList = append(h.interfacesList, APIInterfaces{
		Key:   extractInterfaceNameFromTS(tsRequest),
		Value: tsRequest,
	}, APIInterfaces{
		Key:   extractInterfaceNameFromTS(tsResponse),
		Value: tsResponse,
	})

	// Add to registry
	h.RoutesList = append(h.RoutesList, Route{
		Route:    route.Route,
		Request:  tsRequest,
		Response: tsResponse,
		Method:   method,
		Note:     route.Note,
	})
	return nil
}

// GroupedRoutes organizes routes by their URL path segments for API documentation
func (h *RouteHandler) GroupedRoutes() API {
	const skipSegments = 2  // e.g., skip "api" and "v1" or "v2"
	const groupSegments = 1 // e.g., group by the next segment ("subject")

	groups := make(map[string][]Route)

	for _, route := range h.RoutesList {
		trimmedPath := strings.TrimPrefix(route.Route, "/")
		segments := strings.Split(trimmedPath, "/")

		groupKey := "/"
		if len(segments) > skipSegments {
			end := skipSegments + groupSegments
			if end > len(segments) {
				end = len(segments)
			}
			groupKey = strings.Join(segments[skipSegments:end], "/")
		} else if len(segments) > 0 && segments[0] != "" {
			groupKey = segments[0]
		}

		groups[groupKey] = append(groups[groupKey], route)
	}

	// Sort group keys
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Prepare sorted result
	var groupedRoutes []GroupedRoute
	for _, key := range keys {
		routes := groups[key]
		sort.Slice(routes, func(i, j int) bool {
			return routes[i].Method < routes[j].Method
		})
		groupedRoutes = append(groupedRoutes, GroupedRoute{
			Key:    key,
			Routes: routes,
		})
	}

	return API{
		GroupedRoutes: groupedRoutes,
		Requests:      GetAllRequestInterfaces(h.RoutesList),
		Responses:     GetAllResponseInterfaces(h.RoutesList),
	}
}
