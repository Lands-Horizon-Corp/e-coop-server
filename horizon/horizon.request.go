package horizon

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type HorizonRequest struct {
	Service *echo.Echo
	Log     *HorizonLog // Added field for HorizonLog

	config *HorizonConfig
}

var suspiciousPaths = []string{
	".env", "env.", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml",
	"dockerfile", "Dockerfile", ".git", ".htaccess", ".htpasswd", "backup",
	"secret", "credential", "password", "private", "key", "token", "dump",
	"database", "db", "logs", "debug",
}

func NewHorizonRequest(
	config *HorizonConfig,
	log *HorizonLog, // Pass the HorizonLog instance here
) (*HorizonRequest, error) {
	e := echo.New()
	// Logs
	e.Use(middleware.Logger())

	// Recover from panics
	e.Use(middleware.Recover())

	// Cors
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
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
			"http://localhost:",
			"http://localhost:80",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:4173",
			"http://localhost:8080 ",
		},
		AllowMethods: []string{
			echo.POST,
			echo.PATCH,
			echo.POST,
			echo.DELETE,
			echo.GET,
		},
		AllowHeaders: []string{
			echo.HeaderXCSRFToken,
			echo.HeaderXRequestedWith,
			echo.HeaderAuthorization,
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
		},
		ExposeHeaders: []string{
			echo.HeaderContentLength,
		},
		AllowCredentials: true,
		MaxAge:           60,
	}))

	// Rate limiter
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	// Block suspicious URLs
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := strings.ToLower(c.Request().URL.Path)
			for _, bad := range suspiciousPaths {
				if strings.Contains(path, bad) {
					// Log the suspicious access attempt
					log.LogSecurity(fmt.Sprintf("Suspicious path accessed: %s", path), "warn")
					return c.String(http.StatusForbidden, "Blocked suspicious path")
				}
			}
			// Log the request
			log.LogRequest(fmt.Sprintf("Incoming request: %s %s", c.Request().Method, c.Request().URL.Path), "info")
			return next(c)
		}
	})

	// Health check
	e.GET("/health", func(c echo.Context) error {
		// Log health check request
		log.LogRequest("Health check request", "info")
		return c.String(200, "OK")
	})

	return &HorizonRequest{
		Service: e,
		Log:     log, // Set the logger
		config:  config,
	}, nil
}

func (hr *HorizonRequest) Run() {
	// Start the server in a goroutine
	go func() {
		// Log when the service starts
		hr.Log.LogRequest(fmt.Sprintf("Service started on port %d", hr.config.AppPort), "info")
		hr.Service.Logger.Fatal(
			hr.Service.Start(fmt.Sprintf(":%d", hr.config.AppPort)),
		)
	}()
}
