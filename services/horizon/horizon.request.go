package horizon

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
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
	Route        string `json:"route"`
	Request      string `json:"request,omitempty"`
	Response     string `json:"response,omitempty"`
	RequestType  any
	ResponseType any
	Method       string `json:"method"`
	Note         string `json:"note"`
	Private      bool   `json:"private,omitempty"`
}

// GroupedRoute holds a group of routes under a common key.

type GroupedRoute struct {
	Key    string  `json:"key"`
	Routes []Route `json:"routes"`
}
type APIInterfaces struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type API struct {
	GroupedRoutes []GroupedRoute  `json:"grouped_routes"`
	Requests      []APIInterfaces `json:"requests"`
	Responses     []APIInterfaces `json:"responses"`
}

// HorizonAPIService implements APIService.
type HorizonAPIService struct {
	service        *echo.Echo
	serverPort     int
	metricsPort    int
	clientURL      string
	clientName     string
	routesList     []Route
	interfacesList []APIInterfaces
}

var (
	forbiddenExtensions = []string{
		".env", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml", ".git",
		".htaccess", ".htpasswd", ".backup", ".secret", ".credential", ".password",
		".private", ".key", ".token", ".dump", ".database", ".db", ".logs", ".debug",
		".pem", ".crt", ".cert", ".pfx", ".p12", ".bak", ".swp", ".tmp", ".cache",
		".session", ".sqlite", ".sqlite3", ".mdf", ".ldf", ".rdb", ".ldb", ".log",
		".old", ".orig", ".example", ".sample", ".test", ".spec", ".out", ".core",
	}
	forbiddenSubstrings = []string{
		"dockerfile",
		"credentials",
		"secrets",
		"backup",
		"hidden",
	}
)

// NewHorizonAPIService creates a new API service with sensible defaults.
func NewHorizonAPIService(
	serverPort, metricsPort int,
	clientURL, clientName string,
) APIService {
	e := echo.New()
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("logger.Sync() error: %v\n", err)
		}
	}()

	LoadTemplatesIfExists(e, "public/views/*.html")
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if IsSuspiciousPath(path) {
				return c.String(http.StatusForbidden, "Access forbidden")
			}
			if strings.HasPrefix(strings.ToLower(path), "/.well-known/") {
				return c.String(http.StatusNotFound, "Path not found")
			}
			return next(c)
		}
	})
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

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://ecoop-suite.netlify.app",
			"https://ecoop-suite.com",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:3003",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		},
		AllowHeaders: []string{
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
		},
		ExposeHeaders:    []string{echo.HeaderContentLength},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("REQUEST",
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
				)
			} else {
				logger.Error("REQUEST_ERROR",
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("err", v.Error.Error()),
				)
			}
			return nil
		},
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

	for _, r := range h.routesList {
		if strings.EqualFold(r.Route, route.Route) && strings.EqualFold(r.Method, method) {
			panic(fmt.Sprintf("Route already registered: %s %s", method, route.Route))
		}
	}
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
	if !route.Private {
		tsRequest := TagFormat(route.RequestType)
		tsResponse := TagFormat(route.ResponseType)

		h.interfacesList = append(h.interfacesList, APIInterfaces{
			Key:   ExtractTSInterfaceName(tsRequest),
			Value: tsRequest,
		})
		h.interfacesList = append(h.interfacesList, APIInterfaces{
			Key:   ExtractTSInterfaceName(tsResponse),
			Value: tsResponse,
		})

		h.routesList = append(h.routesList, Route{
			Route:    route.Route,
			Request:  tsRequest,
			Response: tsResponse,
			Method:   method,
			Note:     route.Note,
		})
	}
}

// Run starts the API and metrics servers.
func (h *HorizonAPIService) Run(ctx context.Context) error {

	// New: GET /api/routes returns grouped routes as JSON
	grouped := h.GroupedRoutes()
	h.service.GET("/api/routes", func(c echo.Context) error {
		return c.JSON(http.StatusOK, grouped)
	}).Name = "horizon-routes-json"

	h.service.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "404 - Route not found")
	})

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

func (h *HorizonAPIService) GroupedRoutes() API {
	grouped := make(map[string][]Route)
	interfacesMap := make(map[string]map[string]struct{})
	for _, rt := range h.routesList {
		trimmed := strings.TrimPrefix(rt.Route, "/")
		segments := strings.Split(trimmed, "/")
		key := "/"
		if len(segments) > 0 && segments[0] != "" {
			key = segments[0]
		}
		grouped[key] = append(grouped[key], rt)
		if interfacesMap[key] == nil {
			interfacesMap[key] = make(map[string]struct{})
		}
		// Add request/response interface NAMES
		if rt.Request != "" {
			name := ExtractTSInterfaceName(rt.Request)
			if name != "" {
				interfacesMap[key][name] = struct{}{}
			}
		}
		if rt.Response != "" {
			name := ExtractTSInterfaceName(rt.Response)
			if name != "" {
				interfacesMap[key][name] = struct{}{}
			}
		}
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
		interfaces := make([]string, 0, len(interfacesMap[route]))
		for iface := range interfacesMap[route] {
			interfaces = append(interfaces, iface)
		}
		sort.Strings(interfaces)

		result = append(result, GroupedRoute{
			Key:    route,
			Routes: methodGroup,
		})
	}
	return API{
		GroupedRoutes: result,
		Requests:      GetAllRequestInterfaces(h.routesList),
		Responses:     GetAllResponseInterfaces(h.routesList),
	}
}
