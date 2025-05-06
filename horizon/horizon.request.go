package horizon

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type HorizonRequest struct {
	Service *echo.Echo
	config  *HorizonConfig
	log     *HorizonLog
}

// Compile a regular expression to match suspicious paths
var suspiciousPathPattern = regexp.MustCompile(`(?i)\.(env|yaml|yml|ini|config|conf|xml|git|htaccess|htpasswd|backup|secret|credential|password|private|key|token|dump|database|db|logs|debug)$|dockerfile|Dockerfile`)

func NewHorizonRequest(
	config *HorizonConfig,
	log *HorizonLog,
) (*HorizonRequest, error) {
	e := echo.New()

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := strings.ToLower(c.Request().URL.Path)
			if suspiciousPathPattern.MatchString(path) {
				log.Log(LogEntry{
					Category: CategoryHijack,
					Level:    LevelWarn,
					Message:  fmt.Sprintf("Suspicious path accessed: %s", path),
					Fields: []zap.Field{
						zap.String("method", c.Request().Method),
						zap.String("remote_ip", c.Request().RemoteAddr),
						zap.String("user_agent", c.Request().UserAgent()),
						zap.String("uri", c.Request().RequestURI),
						zap.String("host", c.Request().Host),
						zap.String("referer", c.Request().Referer()),
						zap.String("path", path),
						zap.String("request_id", c.Request().Header.Get(echo.HeaderXRequestID)),
						zap.String("query_params", c.QueryString()),
						zap.String("body", GetRequestBody(c)),
					},
				})
				return c.String(http.StatusForbidden, "Access forbidden")
			}

			if strings.HasPrefix(path, "/.well-known/") {
				return c.String(http.StatusNotFound, "Path not found")
			}
			return next(c)
		}
	})

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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

			log.Log(LogEntry{
				Category: CategoryRequest,
				Level:    LevelInfo,
				Message: fmt.Sprintf("[Request] %s %s | Status: %d | IP: %s | UA: %s | Referer: %s | Latency: %s | Size: %d",
					v.Method,
					v.URI,
					v.Status,
					v.RemoteIP,
					v.UserAgent,
					v.Referer,
					v.Latency,
					v.ResponseSize,
				),
				Fields: []zap.Field{
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("remote_ip", v.RemoteIP),
					zap.String("host", v.Host),
					zap.String("user_agent", v.UserAgent),
					zap.String("referer", v.Referer),
					zap.String("request_id", v.RequestID),
					zap.String("content_length", v.ContentLength),
					zap.Int64("response_size", v.ResponseSize),
					zap.Any("headers", v.Headers),
					zap.Any("query_params", v.QueryParams),
					zap.Any("form_values", v.FormValues),
				},
			})
			return nil
		},
	}))

	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://0.0.0.0",
			"http://0.0.0.0:80",
			"http://0.0.0.0:3000",
			"http://0.0.0.0:3001",
			"http://0.0.0.0:4173",
			"http://0.0.0.0:8080",

			"http://client",
			"http://client:80",
			"http://client:3000",
			"http://client:3001",
			"http://client:4173",
			"http://client:8080",

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

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	return &HorizonRequest{
		Service: e,
		config:  config,
		log:     log,
	}, nil
}

func (hr *HorizonRequest) run() error {
	go func() {
		hr.Service.Logger.Fatal(hr.Service.Start(fmt.Sprintf(":%d", hr.config.AppPort)))
	}()
	return nil
}

func (hr *HorizonRequest) stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := hr.Service.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
	}
	return nil
}
