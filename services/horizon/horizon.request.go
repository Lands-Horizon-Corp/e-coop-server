package horizon

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rotisserie/eris"
	"golang.org/x/time/rate"
)

// APIService defines the interface for an API server.
type APIService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Client() *echo.Echo
	GetRoute() []Route
	RegisterRoute(route Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

// TemplateRenderer implements echo.Renderer for HTML templates.
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	if viewContext, ok := data.(map[string]any); ok {
		viewContext["reverse"] = c.Echo().Reverse
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

// Route describes an API route.
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

// HorizonAPIService implements APIService.
type HorizonAPIService struct {
	service     *echo.Echo
	serverPort  int
	metricsPort int
	clientURL   string
	clientName  string
	routesList  []Route
	certPath    string
	keyPath     string
}

var (
	forbiddenExtensions = []string{
		".env", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml", ".git",
		".htaccess", ".htpasswd", ".backup", ".secret", ".credential", ".password",
		".private", ".key", ".token", ".dump", ".database", ".db", ".logs", ".debug",
	}
	forbiddenSubstrings = []string{
		"dockerfile",
	}
)

// isSuspiciousPath checks if a path is forbidden.

func isSuspiciousPath(path string) bool {
	lower := strings.ToLower(path)
	decoded, _ := url.PathUnescape(lower)
	for _, ext := range forbiddenExtensions {
		if strings.HasSuffix(lower, ext) || strings.HasSuffix(decoded, ext) {
			return true
		}
	}
	for _, substr := range forbiddenSubstrings {
		if strings.Contains(lower, substr) || strings.Contains(decoded, substr) {
			return true
		}
	}
	return false
}

// NewHorizonAPIService creates a new API service with sensible defaults.
func NewHorizonAPIService(
	serverPort, metricsPort int,
	clientURL, clientName string,
) APIService {
	e := echo.New()
	loadTemplatesIfExists(e, "public/views/*.html")

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.BodyLimit("10mb"))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(10),
				Burst:     30,
				ExpiresIn: 5 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "rate limit error"})
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
		},
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if isSuspiciousPath(path) {
				return c.String(http.StatusForbidden, "Access forbidden")
			}
			if strings.HasPrefix(strings.ToLower(path), "/.well-known/") {
				return c.String(http.StatusNotFound, "Path not found")
			}
			return next(c)
		}
	})
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowOriginFunc:  func(string) (bool, error) { return true, nil },
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestedWith, echo.HeaderXCSRFToken, "X-Longitude", "X-Latitude", "Location", "X-Device-Type", "X-User-Agent"},
		ExposeHeaders:    []string{echo.HeaderContentLength},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	e.Use(echoprometheus.NewMiddleware(clientName))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 6,
		Skipper: func(c echo.Context) bool {
			ct := c.Response().Header().Get(echo.HeaderContentType)
			return strings.HasPrefix(ct, "image/") ||
				strings.HasPrefix(ct, "video/") ||
				strings.HasPrefix(ct, "audio/") ||
				strings.HasPrefix(ct, "application/zip") ||
				strings.HasPrefix(ct, "application/pdf")
		},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	return &HorizonAPIService{
		service:     e,
		serverPort:  serverPort,
		metricsPort: metricsPort,
		clientURL:   clientURL,
		clientName:  clientName,
		routesList:  []Route{},
	}
}

// Client returns the Echo instance.
func (h *HorizonAPIService) Client() *echo.Echo { return h.service }

// GetRoute returns all registered routes.
func (h *HorizonAPIService) GetRoute() []Route { return h.routesList }

// RegisterRoute registers a new route and its handler.
func (h *HorizonAPIService) RegisterRoute(route Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	switch method {
	case http.MethodGet:
		h.service.GET(route.Route, callback, m...)
	case http.MethodPost:
		h.service.POST(route.Route, callback, m...)
	case http.MethodPut:
		h.service.PUT(route.Route, callback, m...)
	case http.MethodPatch:
		h.service.PATCH(route.Route, callback, m...)
	case http.MethodDelete:
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

// Run starts the API and metrics servers.
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
			// log error if needed
		}
	}()
	go func() {
		h.service.Logger.Fatal(h.service.Start(fmt.Sprintf(":%d", h.serverPort)))
	}()
	return nil
}

// Stop gracefully shuts down the API server.
func (h *HorizonAPIService) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}

// GroupedRoutes groups routes by their first path segment.
func (h *HorizonAPIService) GroupedRoutes() []GroupedRoute {
	// time.Sleep(5 * time.Second) // Remove or comment out in production.
	grouped := make(map[string][]Route)
	for _, rt := range h.routesList {
		trimmed := strings.TrimPrefix(rt.Route, "/")
		segments := strings.Split(trimmed, "/")
		key := "/"
		if len(segments) > 0 && segments[0] != "" {
			key = segments[0]
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

// loadTemplatesIfExists sets the renderer if templates are found.
func loadTemplatesIfExists(e *echo.Echo, pattern string) {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return
	}
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
}
