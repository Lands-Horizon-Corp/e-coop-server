package horizon

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rotisserie/eris"
	"golang.org/x/time/rate"
)

/*
req.RegisterRoute(horizon.Route{
	Route:    "/sure",
	Method:   "POST",
	Request:  "",
	Response: "string", // or "OK"
	Note:     "Health check endpoint",
}, func(c echo.Context) error {
	return c.String(200, "OK")
})
*/
// APIService defines the interface for an API server with methods for lifecycle control, route registration, and client access.
type APIService interface {
	// Run starts the API service and listens for incoming requests.
	Run(ctx context.Context) error

	// Stop gracefully shuts down the API service.
	Stop(ctx context.Context) error

	// Client returns the underlying Echo instance for advanced customizations.
	Client() *echo.Echo

	// GetRoute returns a list of all registered routes.
	GetRoute() []Route

	RegisterRoute(route Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]any); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

type Route struct {
	Route    string
	Request  string
	Response string
	Method   string
	Note     string
}

// GroupedRoute holds a group of routes under a common key.
type GroupedRoute struct {
	Key    string
	Routes []Route
}

type HorizonAPIService struct {
	service     *echo.Echo
	serverPort  int
	metricsPort int
	clientURL   string
	clientName  string

	routesList []Route

	certPath string
	keyPath  string
}

var suspiciousPathPattern = regexp.MustCompile(`(?i)\.(env|yaml|yml|ini|config|conf|xml|git|htaccess|htpasswd|backup|secret|credential|password|private|key|token|dump|database|db|logs|debug)$|dockerfile|Dockerfile`)

func NewHorizonAPIService(
	serverPort int,
	metricsPort int,
	clientURL string,
	clientName string,
) APIService {
	service := echo.New()
	loadTemplatesIfExists(service, "public/views/*.html")

	service.Pre(middleware.RemoveTrailingSlash())
	service.Use(middleware.BodyLimit("10mb"))
	service.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// 5. Rate limiting
	service.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := strings.ToLower(c.Request().URL.Path)
			if suspiciousPattern := suspiciousPathPattern.MatchString(path); suspiciousPattern {
				return c.String(http.StatusForbidden, "Access forbidden")
			}
			if strings.HasPrefix(path, "/.well-known/") {
				return c.String(http.StatusNotFound, "Path not found")
			}
			return next(c)
		}
	})

	service.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowOriginFunc: func(origin string) (bool, error) {
			return true, nil
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}, AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestedWith,
			echo.HeaderXCSRFToken,
			"X-Longitude",
			"X-Latitude",
			"Location",
			"X-Device-Type",
			"X-User-Agent",
		}, ExposeHeaders: []string{echo.HeaderContentLength},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	// 9. Metrics middleware
	service.Use(echoprometheus.NewMiddleware(clientName))

	service.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 6,
	}))
	service.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	return &HorizonAPIService{
		service:     service,
		serverPort:  serverPort,
		metricsPort: metricsPort,
		clientURL:   clientURL,
		clientName:  clientName,
		routesList:  []Route{},
	}
}

// Client implements APIService.
func (h *HorizonAPIService) Client() *echo.Echo {
	return h.service
}

// GetRoute implements APIService.
func (h *HorizonAPIService) GetRoute() []Route {
	return h.routesList
}

func (h *HorizonAPIService) RegisterRoute(route Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	switch method {
	case "GET":
		h.service.GET(route.Route, callback, m...)
	case "POST":
		h.service.POST(route.Route, callback, m...)
	case "PUT":
		h.service.PUT(route.Route, callback, m...)
	case "PATCH":
		h.service.PATCH(route.Route, callback, m...)
	case "DELETE":
		h.service.DELETE(route.Route, callback, m...)
	default:
		panic(fmt.Sprintf("Unsupported HTTP method: %s", method))
	}
	h.routesList = append(h.routesList, Route{
		Route:    route.Route,
		Request:  route.Request,
		Response: route.Response,
		Method:   method,
		Note:     route.Note,
	})
}

// Run implements APIService.
func (h *HorizonAPIService) Run(ctx context.Context) error {

	h.service.GET("/routes", func(c echo.Context) error {
		return c.Render(http.StatusOK, "routes.html", map[string]any{
			"routes": h.GroupedRoutes(),
		})

	}).Name = "horizon-routes"

	h.service.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "404 - Route not found")
	})
	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(fmt.Sprintf(":%d", h.metricsPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// skip
		}
	}()
	go func() {
		h.service.Logger.Fatal(h.service.Start(
			fmt.Sprintf(":%d", h.serverPort),
		))

	}()
	return nil
}

// Stop implements APIService.
func (h *HorizonAPIService) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}
func (h *HorizonAPIService) GroupedRoutes() []GroupedRoute {
	time.Sleep(5 * time.Second) // Simulate delay, can be removed if not needed.

	grouped := make(map[string][]Route)
	for _, rt := range h.routesList {
		trimmed := strings.TrimPrefix(rt.Route, "/")
		segments := strings.Split(trimmed, "/")
		var key string
		if len(segments) > 0 && segments[0] != "" {
			key = segments[0]
		} else {
			key = "/"
		}
		grouped[key] = append(grouped[key], rt)
	}

	routePaths := make([]string, 0, len(grouped))
	for route := range grouped {
		routePaths = append(routePaths, route)
	}
	sort.Strings(routePaths)

	result := make([]GroupedRoute, 0, len(routePaths))
	for _, route := range routePaths {
		methodGroup := grouped[route]
		sort.Slice(methodGroup, func(i, j int) bool {
			return methodGroup[i].Method < methodGroup[j].Method
		})
		result = append(result, GroupedRoute{
			Key:    route,
			Routes: methodGroup,
		})
	}
	return result
}

func loadTemplatesIfExists(service *echo.Echo, pattern string) {
	// Check if any files match the pattern
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		// No templates found, skip setting renderer
		return
	}

	// Parse templates if found
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
	service.Renderer = renderer
}
