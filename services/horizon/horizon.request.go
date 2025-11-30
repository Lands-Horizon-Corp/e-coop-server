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

// NewHorizonAPIService creates a new API service with comprehensive security middleware.
// The secured flag indicates if the server is in production mode (HTTPS enforced)
func NewHorizonAPIService(
	cache CacheService,
	serverPort, metricsPort int,
	clientURL, clientName string,
	secured bool,
) APIService {
	//===== ECHO INSTANCE AND LOGGER SETUP =====
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

	//===== CORS ORIGINS CONFIGURATION =====
	// Production domains that are allowed to make requests
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

	// Add development origins when not in secured mode
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

	// Extract hostnames from origins for host validation
	allowedHosts := make([]string, 0, len(origins))
	for _, origin := range origins {
		hostname := strings.TrimPrefix(origin, "https://")
		hostname = strings.TrimPrefix(hostname, "http://")
		allowedHosts = append(allowedHosts, hostname)
	}

	//===== BASIC MIDDLEWARE SETUP =====
	// Panic recovery middleware
	e.Use(middleware.Recover())

	// Remove trailing slashes from URLs
	e.Pre(middleware.RemoveTrailingSlash())

	// Force HTTPS redirect in production
	if secured {
		e.Pre(middleware.HTTPSRedirect())
	}

	//===== HOST VALIDATION MIDDLEWARE =====
	// Only allow requests from approved domains
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			host := handlers.GetHost(c)
			if slices.Contains(allowedHosts, host) {
				return next(c)
			}
			return c.String(http.StatusForbidden, "Host not allowed")
		}
	})

	//===== HTTP METHOD RESTRICTION MIDDLEWARE =====
	// Only allow standard HTTP methods, reject unusual ones
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

	//===== IP FIREWALL MIDDLEWARE =====
	// Check blocked IPs from Redis cache
	// Blocks IPs from HaGeZi blocklists and manually flagged malicious IPs
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := handlers.GetClientIP(c)
			if net.ParseIP(clientIP) == nil {
				logger.Warn("Invalid IP format detected",
					zap.String("raw_ip", clientIP),
					zap.String("user_agent", handlers.GetUserAgent(c)),
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

	//===== SUSPICIOUS PATH DETECTION MIDDLEWARE =====
	// Detect and block injection attempts, directory traversal, etc.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := handlers.GetPath(c)

			suspiciousCacheKey := "suspicious_path:" + path

			cachedResult, err := cache.Get(c.Request().Context(), suspiciousCacheKey)
			if err == nil && cachedResult != nil {
				logger.Warn("Suspicious path blocked (cached)",
					zap.String("ip", handlers.GetClientIP(c)),
					zap.String("path", path),
				)
				return c.String(http.StatusForbidden, "Access forbidden")
			}
			if handlers.IsSuspiciousPath(path) {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := cache.Set(ctx, suspiciousCacheKey, []byte("blocked"), 5*time.Minute); err != nil {
						logger.Error("Failed to cache suspicious path",
							zap.String("path", path),
							zap.Error(err),
						)
					}
				}()
				logger.Warn("Suspicious path blocked",
					zap.String("ip", handlers.GetClientIP(c)),
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

	//===== REQUEST SIZE LIMIT MIDDLEWARE =====
	// Limit request body size to prevent DoS attacks
	e.Use(middleware.BodyLimit("10mb"))

	//===== SECURITY HEADERS MIDDLEWARE =====
	if secured {
		// Comprehensive security headers for production
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

		// Extended security headers for production (Permissions Policy, CT, CORP, etc.)
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
		// Relaxed security headers for development environment
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN", // More lenient for development
			HSTSMaxAge:            0,            // Disable HSTS in development
			HSTSExcludeSubdomains: true,         // OK for development
			HSTSPreloadEnabled:    false,
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
				"img-src 'self' data: https: http:; " +
				"connect-src 'self' ws: wss: http: https:; " +
				"frame-src 'self';",
		}))
	}

	//===== RATE LIMITING MIDDLEWARE =====
	// Prevent abuse with IP + User-Agent based rate limiting
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(20),
				Burst:     100,
				ExpiresIn: 5 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			ip := handlers.GetClientIP(ctx)
			userAgent := handlers.GetUserAgent(ctx)
			return fmt.Sprintf("%s:%s", ip, userAgent), nil
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

	//===== CORS MIDDLEWARE =====
	// Configure Cross-Origin Resource Sharing for allowed domains
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
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

	//===== CORS PREFLIGHT DEBUGGING MIDDLEWARE =====
	// Enhanced CORS preflight handling with detailed logging
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				origin := c.Request().Header.Get("Origin")
				logger.Info("CORS Preflight Request",
					zap.String("origin", origin),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Header.Get("Access-Control-Request-Method")),
					zap.String("headers", c.Request().Header.Get("Access-Control-Request-Headers")),
				)

				// Verify origin is allowed before setting headers
				if slices.Contains(origins, origin) {
					c.Response().Header().Set("Access-Control-Allow-Origin", origin)
					c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
					c.Response().Header().Set("Access-Control-Max-Age", "3600")
				} else {
					logger.Warn("CORS request from unauthorized origin",
						zap.String("origin", origin),
						zap.String("path", c.Request().URL.Path),
					)
				}

				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	})

	//===== REQUEST LOGGING MIDDLEWARE =====
	// Log all incoming requests with status and error details
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

	//===== COMPRESSION MIDDLEWARE =====
	// Gzip compression for text-based responses (skips binary content)
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

	//===== TIMEOUT MIDDLEWARE =====
	// Prevent long-running requests from consuming resources
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

	//===== DEFAULT ENDPOINTS =====
	// Health check endpoint for monitoring
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Root welcome endpoint
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
