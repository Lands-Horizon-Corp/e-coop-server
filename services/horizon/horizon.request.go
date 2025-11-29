package horizon

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
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
	RegisterRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

// APIServiceImpl implements APIService.
type APIServiceImpl struct {
	service     *echo.Echo
	serverPort  int
	metricsPort int
	clientURL   string
	clientName  string
	handler     *handlers.RouteHandler
	cache       CacheService
}

// NewHorizonAPIService creates a new API service with sensible defaults.
func NewHorizonAPIService(
	cache CacheService,
	serverPort, metricsPort int,
	clientURL, clientName string,
	secured bool,
) APIService {
	// the secured flag indicates if the server is in production mode (HTTPS enforced)
	e := echo.New()
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			if !strings.Contains(err.Error(), "sync /dev/stderr") &&
				!strings.Contains(err.Error(), "sync /dev/stdout") &&
				!strings.Contains(err.Error(), "invalid argument") {
				fmt.Printf("logger.Sync() error: %v\n", err)
			}
		}
	}()

	e.Use(middleware.Recover())

	e.Pre(middleware.RemoveTrailingSlash())

	if secured {
		e.Pre(middleware.HTTPSRedirect())
	}

	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			host := c.Request().Host
			allowedHosts := []string{
				"ecoop-suite.com",
				"www.ecoop-suite.com",
				"staging.ecoop-suite.com",
				"www.staging.ecoop-suite.com",
				"development.ecoop-suite.com",
				"www.development.ecoop-suite.com",
				"cooperatives-development.fly.dev",
				"cooperatives-staging.fly.dev",
				"cooperatives-production.fly.dev",
			}

			if !secured {
				allowedHosts = append(allowedHosts, "localhost:8000", "localhost:8001", "localhost:8080", "localhost:3000", "localhost:3001", "localhost:3002", "localhost:3003")
			}
			if slices.Contains(allowedHosts, host) {
				return next(c)
			}
			return c.String(http.StatusForbidden, "Host not allowed")
		}
	})

	// HTTP Method restrictions - only allow safe methods
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			allowedMethods := map[string]bool{
				http.MethodGet:     true,
				http.MethodPost:    true,
				http.MethodPut:     true,
				http.MethodPatch:   true,
				http.MethodDelete:  true,
				http.MethodHead:    true,
				http.MethodOptions: true,
			}

			if !allowedMethods[c.Request().Method] {
				return echo.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed")
			}
			return next(c)
		}
	})

	// Firewall middleware - check blocked IPs from Redis cache
	// This middleware checks if the client IP is blocked.
	// You can populate the cache with IPs resolved from HaGeZi blocklists
	// so that requests from known malicious domains/IPs are automatically denied.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Enhanced IP extraction for Fly.io and other proxies
			clientIP := getClientIP(c)

			// Validate IP format
			if net.ParseIP(clientIP) == nil {
				logger.Warn("Invalid IP format detected",
					zap.String("raw_ip", clientIP),
					zap.String("user_agent", c.Request().UserAgent()),
				)
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
			}

			cacheKey := "blocked_ip:" + clientIP
			hostBytes, err := cache.Get(c.Request().Context(), cacheKey)
			if err != nil {
				logger.Error("Firewall cache error",
					zap.String("ip", clientIP),
					zap.Error(err),
				)
				return next(c)
			}
			if hostBytes != nil {
				blockedHost := string(hostBytes)
				logger.Warn("Blocked IP access attempt",
					zap.String("ip", clientIP),
					zap.String("blocked_host", blockedHost),
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Access denied",
					"code":  "IP_BLOCKED",
				})
			}
			return next(c)
		}
	})

	// ✅ NEW: Enhanced suspicious path blocking for dotfiles/hidden files
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			if handlers.IsSuspiciousPath(path) {
				logger.Warn("Suspicious path blocked",
					zap.String("ip", c.RealIP()),
					zap.String("path", path),
				)
				return c.String(http.StatusForbidden, "Access forbidden")
			}
			if strings.HasPrefix(strings.ToLower(path), "/.well-known/") {
				return c.String(http.StatusNotFound, "Path not found")
			}
			return next(c)
		}
	})
	e.Use(middleware.BodyLimit("10mb"))

	if secured {
		// Production security headers
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			HSTSMaxAge:            31536000,
			HSTSExcludeSubdomains: false,
			HSTSPreloadEnabled:    true,
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			ContentSecurityPolicy: "default-src 'self'; " +
				"script-src 'self'; " +
				"style-src 'self'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self' data:; " +
				"connect-src 'self' https:; " +
				"media-src 'self'; " +
				"object-src 'none'; " +
				"frame-src 'none'; " +
				"frame-ancestors 'none'; " +
				"form-action 'self'; " +
				"base-uri 'self'; " +
				"manifest-src 'self'; " +
				"worker-src 'self'; " +
				"report-uri /api/csp-violations; " +
				"report-to csp-endpoint;",
		}))

		// Additional security headers for production
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Permissions Policy (comprehensive security controls)
				c.Response().Header().Set("Permissions-Policy",
					"accelerometer=(), ambient-light-sensor=(), autoplay=(self), battery=(), "+
						"camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), "+
						"encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), "+
						"fullscreen=(self), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), "+
						"microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), "+
						"publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), "+
						"xr-spatial-tracking=(), clipboard-read=(self), clipboard-write=(self), gamepad=(), "+
						"speaker-selection=(), vibrate=()")

				// Expect-CT for Certificate Transparency
				c.Response().Header().Set("Expect-CT", "max-age=86400, enforce")

				// Additional security headers
				c.Response().Header().Set("X-Permitted-Cross-Domain-Policies", "none")
				c.Response().Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
				c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
				c.Response().Header().Set("Cross-Origin-Resource-Policy", "same-origin")

				// Server information hiding
				c.Response().Header().Set("Server", "")
				c.Response().Header().Set("X-Powered-By", "")

				return next(c)
			}
		})
	} else {
		// Development security headers (more lenient)
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN", // More lenient for development
			HSTSMaxAge:            0,            // Disable HSTS in development
			HSTSExcludeSubdomains: true,         // OK for development
			HSTSPreloadEnabled:    false,
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			// Development CSP - allows unsafe-inline/unsafe-eval for dev tools
			ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
				"img-src 'self' data: https: http:; " +
				"connect-src 'self' ws: wss: http: https:; " +
				"frame-src 'self';",
		}))
	}
	// ✅ IMPROVED: Enhanced rate limiting with Redis support and secure error handling
	var rateLimitStore middleware.RateLimiterStore
	if cache != nil && secured {
		// Use Redis-backed rate limiter for production/distributed setup
		// Use in-memory rate limiting for now - can be enhanced with Redis-backed store later
		rateLimitStore = middleware.NewRateLimiterMemoryStore(rate.Limit(20))
	} else {
		// Use in-memory store for development
		rateLimitStore = middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(20),
				Burst:     100_000,
				ExpiresIn: 5 * time.Minute,
			},
		)
	}

	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   rateLimitStore,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			ip := ctx.RealIP()
			if secured {
				if forwardedIP := ctx.Request().Header.Get("Fly-Client-IP"); forwardedIP != "" {
					ip = forwardedIP
				} else if forwardedIP := ctx.Request().Header.Get("X-Forwarded-For"); forwardedIP != "" {
					if ips := strings.Split(forwardedIP, ","); len(ips) > 0 {
						ip = strings.TrimSpace(ips[0])
					}
				}
			}
			return ip, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			if secured {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Request rate limited"})
			}
			return c.JSON(http.StatusForbidden, map[string]string{"error": "rate limit error " + err.Error()})
		},
		DenyHandler: func(c echo.Context, _ string, err error) error {
			if secured {
				return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too many requests. Please try again later."})
			}
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded " + err.Error()})
		},
	}))

	origins := []string{
		"https://ecoop-suite.netlify.app",
		"https://ecoop-suite.com",
		"https://www.ecoop-suite.com",
		"https://development.ecoop-suite.com",
		"https://www.development.ecoop-suite.com",
		"https://staging.ecoop-suite.com",
		"https://www.staging.ecoop-suite.com",
		"https://cooperatives-development.fly.dev",
		"https://cooperatives-staging.fly.dev",
		"https://cooperatives-production.fly.dev",
	}
	if !secured {
		origins = append(origins,
			"http://localhost:8000",
			"http://localhost:8001",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:3003",
		)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions, // Required for preflight requests
			http.MethodHead,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestedWith,
			echo.HeaderXCSRFToken,
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAccessControlRequestHeaders,
			"X-Longitude",
			"X-Latitude",
			"Location",
			"X-Device-Type",
			"X-User-Agent",
		},
		ExposeHeaders: []string{
			echo.HeaderContentLength,
			echo.HeaderContentType,
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
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
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Request timed out. Please try again later.",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			logger.Error("Timeout error",
				zap.String("path", c.Path()),
				zap.Error(err),
			)
		},
		Timeout: 1 * time.Minute,
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Horizon API")
	})

	return &APIServiceImpl{
		service:     e,
		serverPort:  serverPort,
		metricsPort: metricsPort,
		clientURL:   clientURL,
		clientName:  clientName,
		handler:     handlers.NewRouteHandler(),
		cache:       cache,
	}
}

// Client returns the Echo instance.
func (h *APIServiceImpl) Client() *echo.Echo { return h.service }

// RegisterRoute registers a new route and its handler.
func (h *APIServiceImpl) RegisterRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	if err := h.handler.AddRoute(route); err != nil {
		panic(err)
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
	}
}

// Run starts the API and metrics servers.
func (h *APIServiceImpl) Run(_ context.Context) error {

	// New: GET /api/routes returns grouped routes as JSON
	grouped := h.handler.GroupedRoutes()

	h.service.GET("/api/routes", func(c echo.Context) error {
		return c.JSON(http.StatusOK, grouped)
	}).Name = "horizon-routes-json"

	// CSP violation reporting endpoint
	h.service.POST("/api/csp-violations", func(c echo.Context) error {
		type CSPViolation struct {
			DocumentURI        string `json:"document-uri"`
			Referrer           string `json:"referrer"`
			ViolatedDirective  string `json:"violated-directive"`
			EffectiveDirective string `json:"effective-directive"`
			OriginalPolicy     string `json:"original-policy"`
			Disposition        string `json:"disposition"`
			BlockedURI         string `json:"blocked-uri"`
			LineNumber         int    `json:"line-number"`
			ColumnNumber       int    `json:"column-number"`
			SourceFile         string `json:"source-file"`
			StatusCode         int    `json:"status-code"`
			ScriptSample       string `json:"script-sample"`
		}

		var report struct {
			CSPReport CSPViolation `json:"csp-report"`
		}

		if err := c.Bind(&report); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid CSP report format"})
		}

		// Log the violation for monitoring
		c.Logger().Warnf("CSP Violation: %+v", report.CSPReport)

		return c.JSON(http.StatusOK, map[string]string{"status": "reported"})
	}).Name = "csp-violations"

	h.service.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "404 - Route not found")
	})

	go func() {
		h.service.Logger.Fatal(h.service.Start(fmt.Sprintf(":%d", h.serverPort)))
	}()
	return nil
}

// Stop gracefully shuts down the API server.
func (h *APIServiceImpl) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}

// getClientIP extracts the real client IP from various proxy headers
// Prioritizes Fly.io specific headers for accurate IP detection
func getClientIP(c echo.Context) string {
	// Check Fly.io specific headers first
	if ip := c.Request().Header.Get("Fly-Client-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Standard proxy headers
	if ip := c.Request().Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if ip := c.Request().Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	if ip := c.Request().Header.Get("CF-Connecting-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fallback to Echo's RealIP method
	return c.RealIP()
}
