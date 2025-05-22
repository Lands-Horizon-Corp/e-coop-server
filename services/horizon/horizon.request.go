package horizon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rotisserie/eris"
	echoSwagger "github.com/swaggo/echo-swagger"
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

const (
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
	Cyan   = "\033[36m"
)

type Route struct {
	Route    string
	Request  string
	Response string
	Method   string
	Note     string
}

type HorizonAPIService struct {
	service     *echo.Echo
	serverPort  int
	metricsPort int
	clientURL   string
	clientName  string

	routesList []Route
}

var suspiciousPathPattern = regexp.MustCompile(`(?i)\.(env|yaml|yml|ini|config|conf|xml|git|htaccess|htpasswd|backup|secret|credential|password|private|key|token|dump|database|db|logs|debug)$|dockerfile|Dockerfile`)

func NewHorizonAPIService(
	serverPort int,
	metricsPort int,
	clientURL string,
	clientName string,
) APIService {
	service := echo.New()

	service.Pre(middleware.RemoveTrailingSlash())

	service.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	service.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogURIPath:       true,
		LogStatus:        true,
		LogMethod:        true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogUserAgent:     true,
		LogReferer:       true,
		LogLatency:       true,
		LogRequestID:     true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogHeaders:       []string{"Authorization", "Content-Type"},
		LogQueryParams:   []string{"*"},
		LogFormValues:    []string{"*"},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			return nil
		},
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
		AllowOrigins: []string{
			"http://0.0.0.0",
			"http://0.0.0.0:80",
			"http://0.0.0.0:3000",
			"http://0.0.0.0:3001",
			"http://0.0.0.0:4173",
			"http://0.0.0.0:8080",

			// Client Docker
			"http://client",
			"http://client:80",
			"http://client:3000",
			"http://client:3001",
			"http://client:4173",
			"http://client:8080",

			// Localhost
			"http://localhost",
			"http://localhost:80",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:4173",
			"http://localhost:8080",
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			clientURL,
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
		}, ExposeHeaders: []string{echo.HeaderContentLength},
		AllowCredentials: true, // must be true if the client sends cookies/auth
		MaxAge:           3600,
	}))

	// 9. Metrics middleware
	service.Use(echoprometheus.NewMiddleware(clientName))

	service.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
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

// RegisterRouteDELETE implements APIService.
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

	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(fmt.Sprintf(":%d", h.metricsPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// skip
		}
	}()
	go func() {
		h.service.GET("/swagger/*", echoSwagger.WrapHandler)
		h.service.Logger.Fatal(h.service.Start(
			fmt.Sprintf(":%d", h.serverPort),
		))
	}()
	h.PrintGroupedRoute()
	return nil
}

// Stop implements APIService.
func (h *HorizonAPIService) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}
func (h *HorizonAPIService) PrintGroupedRoute() {
	time.Sleep(5 * time.Second)

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

	fmt.Printf("\n\n================== API ROUTES ==================\n\n")

	for _, route := range routePaths {
		methodGroup := grouped[route]

		sort.Slice(methodGroup, func(i, j int) bool {
			return methodGroup[i].Method < methodGroup[j].Method
		})

		fmt.Printf("🔹 %s Route Group %s: [%d] %s\n", Cyan, route, len(methodGroup), Reset)
		fmt.Println("------------------------------------------------")

		for _, rt := range methodGroup {
			color := ""
			switch rt.Method {
			case "GET":
				color = Green
			case "POST":
				color = Blue
			case "PUT":
				color = Yellow
			case "DELETE":
				color = Red
			default:
				color = Reset
			}

			req := rt.Request
			if req == "" {
				req = "No Request"
			}
			res := rt.Response
			if res == "" {
				res = "No Response"
			}

			fmt.Printf("\t%s⇒ %-6s\t%s%s%s%s\n", color, rt.Method, Reset, color, rt.Route, Reset)

			fmt.Printf("\t\033[36mRequest ⇒\033[0m   \t%s\n", req)
			fmt.Printf("\t\033[36mResponse ⇒\033[0m  \t%s\n", res)

			if rt.Note != "" {
				fmt.Printf("\t\033[2mNote         \t%s\n\n\033[0m", rt.Note)
			} else {
				fmt.Printf("\n")
			}
		}

		fmt.Printf("------------------------------------------------\n")
	}
}
