// Package horizon provides a comprehensive, security-focused HTTP API service
// with Redis-backed middleware for production deployment on Fly.io
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
)

// SecurityHeaderConfig contains configuration for security headers
type SecurityHeaderConfig struct {
	ContentTypeNosniff    string
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreloadEnabled    bool
	ReferrerPolicy        string
	ContentSecurityPolicy string
}

// ExtendedSecurityHeaders contains additional production security headers
type ExtendedSecurityHeaders struct {
	PermissionsPolicy             string
	ExpectCT                      string
	XPermittedCrossDomainPolicies string
	CrossOriginEmbedderPolicy     string
	CrossOriginOpenerPolicy       string
	CrossOriginResourcePolicy     string
}

// APIServiceImpl implements APIService with comprehensive security middleware stack.
// This implementation provides enterprise-grade security features including:
// - Redis-backed rate limiting and IP blocking
// - Comprehensive security headers
// - Request validation and suspicious path detection
// - Environment-aware configuration (dev/prod)
type APIServiceImpl struct {
	service    *echo.Echo             // Echo web framework instance
	serverPort int                    // HTTP server port
	handler    *handlers.RouteHandler // Route management handler
	cache      CacheService           // Redis cache service for security features
}

// APIService defines the interface for a secure, production-ready HTTP API server
// with comprehensive middleware and Redis-backed security features.
type APIService interface {
	// Run starts the API server with all configured middleware
	Run(ctx context.Context) error

	// Stop gracefully shuts down the API server
	Stop(ctx context.Context) error

	// Client returns the underlying Echo instance for advanced configuration
	Client() *echo.Echo

	// RegisterRoute adds a new route with optional middleware
	RegisterWebRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

// getProductionSecurityConfig returns security header configuration for production
func getProductionSecurityConfig() SecurityHeaderConfig {
	return SecurityHeaderConfig{
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,     // Include all subdomains
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; " +
			"script-src 'self' 'nonce-{random}'; " + // Strict script sources with nonce support
			"style-src 'self' 'nonce-{random}'; " + // Strict style sources with nonce support
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
			"sandbox allow-scripts allow-same-origin allow-forms; " + // Add sandbox directive
			"require-trusted-types-for 'script'; " + // Trusted Types API
			"report-uri /api/csp-violations; " +
			"report-to csp-endpoint;",
	}
}

// getDevelopmentSecurityConfig returns relaxed security headers for development
func getDevelopmentSecurityConfig() SecurityHeaderConfig {
	return SecurityHeaderConfig{
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN", // More lenient for development
		HSTSMaxAge:            0,            // Disable HSTS in development
		HSTSIncludeSubdomains: false,        // OK for development
		HSTSPreloadEnabled:    false,
		ReferrerPolicy:        "no-referrer-when-downgrade", // Allow HTTP in development
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval' http: https:; " +
			"img-src 'self' data: https: http:; " +
			"connect-src 'self' ws: wss: http: https:; " +
			"frame-src 'self' http: https:; " +
			"form-action 'self' http: https:;", // Allow HTTP forms in development
	}
}

// getExtendedSecurityHeaders returns additional production security headers
func getExtendedSecurityHeaders() ExtendedSecurityHeaders {
	return ExtendedSecurityHeaders{
		PermissionsPolicy: "accelerometer=(), ambient-light-sensor=(), autoplay=(), battery=(), " +
			"camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), " +
			"encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), " +
			"fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), " +
			"microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), " +
			"publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), " +
			"xr-spatial-tracking=(), clipboard-read=(), clipboard-write=(), gamepad=(), " +
			"speaker-selection=(), vibrate=()",
		ExpectCT:                      "max-age=86400, enforce",
		XPermittedCrossDomainPolicies: "none",
		CrossOriginEmbedderPolicy:     "require-corp",
		CrossOriginOpenerPolicy:       "same-origin",
		CrossOriginResourcePolicy:     "same-origin",
	}
}

// applyExtendedSecurityHeaders applies additional security headers to the response
func applyExtendedSecurityHeaders(c echo.Context, headers ExtendedSecurityHeaders) {
	// Permissions Policy (comprehensive security controls)
	c.Response().Header().Set("Permissions-Policy", headers.PermissionsPolicy)

	// Expect-CT for Certificate Transparency
	c.Response().Header().Set("Expect-CT", headers.ExpectCT)

	// Additional security headers
	c.Response().Header().Set("X-Permitted-Cross-Domain-Policies", headers.XPermittedCrossDomainPolicies)
	c.Response().Header().Set("Cross-Origin-Embedder-Policy", headers.CrossOriginEmbedderPolicy)
	c.Response().Header().Set("Cross-Origin-Opener-Policy", headers.CrossOriginOpenerPolicy)
	c.Response().Header().Set("Cross-Origin-Resource-Policy", headers.CrossOriginResourcePolicy)

	// Server information hiding
	c.Response().Header().Set("Server", "")
	c.Response().Header().Set("X-Powered-By", "")
}

// SecurityHeadersMiddleware applies security headers as final middleware to ensure consistency
func SecurityHeadersMiddleware(secured bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Execute the next handler first
			err := next(c)

			// Apply security headers after response is generated
			if secured {
				// Production security headers configuration
				securityConfig := getProductionSecurityConfig()
				extendedHeaders := getExtendedSecurityHeaders()

				// Apply core security headers
				c.Response().Header().Set("X-Content-Type-Options", securityConfig.ContentTypeNosniff)
				c.Response().Header().Set("X-Frame-Options", securityConfig.XFrameOptions)
				c.Response().Header().Set("Referrer-Policy", securityConfig.ReferrerPolicy)
				c.Response().Header().Set("Content-Security-Policy", securityConfig.ContentSecurityPolicy)

				// Apply HSTS headers
				if securityConfig.HSTSMaxAge > 0 {
					hstsValue := fmt.Sprintf("max-age=%d", securityConfig.HSTSMaxAge)
					if securityConfig.HSTSIncludeSubdomains {
						hstsValue += "; includeSubDomains"
					}
					if securityConfig.HSTSPreloadEnabled {
						hstsValue += "; preload"
					}
					c.Response().Header().Set("Strict-Transport-Security", hstsValue)
				}

				// Apply extended security headers
				applyExtendedSecurityHeaders(c, extendedHeaders)
			} else {
				// Development security headers configuration
				securityConfig := getDevelopmentSecurityConfig()

				// Apply basic security headers for development
				c.Response().Header().Set("X-Content-Type-Options", securityConfig.ContentTypeNosniff)
				c.Response().Header().Set("X-Frame-Options", securityConfig.XFrameOptions)
				c.Response().Header().Set("Referrer-Policy", securityConfig.ReferrerPolicy)
				c.Response().Header().Set("Content-Security-Policy", securityConfig.ContentSecurityPolicy)
			}

			return err
		}
	}
}

// NewHorizonAPIService creates a new API service with comprehensive security middleware.
// This function sets up a production-ready Echo server with 15+ security layers including:
// - Host validation and HTTPS enforcement
// - IP firewall with Redis-backed blocklists
// - Sophisticated path injection detection
// - Redis-distributed rate limiting
// - Comprehensive security headers
// - CORS configuration with origin validation
//
// Parameters:
//   - cache: Redis cache service for security features (IP blocking, rate limiting, path caching)
//   - serverPort: HTTP server port number
//   - secured: Production mode flag (enables HTTPS redirect and strict security headers)
//
// Returns: Configured APIService ready for production deployment
func NewHorizonAPIService(
	cache CacheService,
	serverPort int,
	secured bool,
) APIService {

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

	// Production domains
	origins := []string{
		// Primary production domains
		"https://ecoop-suite.netlify.app",
		"https://ecoop-suite.com",
		"https://www.ecoop-suite.com",

		// Development and staging environments
		"https://development.ecoop-suite.com",
		"https://www.development.ecoop-suite.com",
		"https://staging.ecoop-suite.com",
		"https://www.staging.ecoop-suite.com",

		// Fly.io deployment domains
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

	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	// Force HTTPS redirect in production
	if secured {
		e.Pre(middleware.HTTPSRedirect())
	}

	// Host validation middleware
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			host := handlers.GetHost(c)

			// Check if the host is in our allowlist
			if slices.Contains(allowedHosts, host) {
				return next(c)
			}

			// Reject requests from unauthorized hosts
			return c.String(http.StatusForbidden, "Host not allowed")
		}
	})

	// HTTP method restriction middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Define allowed HTTP methods for API operations
			allowedMethods := map[string]bool{
				http.MethodGet:     true, // Read operations
				http.MethodPost:    true, // Create operations
				http.MethodPut:     true, // Update/replace operations
				http.MethodPatch:   true, // Partial update operations
				http.MethodDelete:  true, // Delete operations
				http.MethodHead:    true, // Header-only requests
				http.MethodOptions: true, // CORS preflight requests
			}

			// Reject requests with unauthorized HTTP methods
			if !allowedMethods[c.Request().Method] {
				return echo.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed")
			}

			return next(c)
		}
	})

	// IP firewall middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract and validate client IP address
			clientIP := handlers.GetClientIP(c)

			// Validate IP format to prevent injection attacks
			if net.ParseIP(clientIP) == nil {
				logger.Warn("Invalid IP format detected - potential attack",
					zap.String("raw_ip", clientIP),
					zap.String("user_agent", handlers.GetUserAgent(c)),
				)
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
			}

			// Check if IP is blocked
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
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					timestamp := float64(time.Now().Unix())

					// Track blocked IP attempts in sorted set
					attemptKey := "blocked_attempts:" + clientIP
					if err := cache.ZAdd(ctx, attemptKey, timestamp, fmt.Sprintf("%s:%d", c.Request().URL.Path, time.Now().Unix())); err != nil {
						logger.Debug("Failed to track blocked IP attempt",
							zap.String("ip", clientIP),
							zap.Error(err))
					}

					// Also track in global blocked IPs registry for analytics
					if err := cache.ZAdd(ctx, "blocked_ips_registry", timestamp, clientIP); err != nil {
						logger.Debug("Failed to update blocked IPs registry",
							zap.String("ip", clientIP),
							zap.Error(err))
					}

					// Cleanup old attempts (keep last 7 days)
					sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()
					if _, err := cache.ZRemRangeByScore(ctx, attemptKey, "0", fmt.Sprintf("%d", sevenDaysAgo)); err != nil {
						logger.Debug("Failed to cleanup old blocked attempts",
							zap.String("ip", clientIP),
							zap.Error(err))
					}
				}()
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

	// Suspicious path detection middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := handlers.GetPath(c)
			clientIP := handlers.GetClientIP(c)

			// Generate cache key for suspicious path detection
			suspiciousCacheKey := "suspicious_path:" + path

			cachedResult, err := cache.Get(c.Request().Context(), suspiciousCacheKey)
			if err == nil && cachedResult != nil {
				// Track repeated suspicious path attempt using ZADD
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					// Track suspicious path attempts per IP
					timestamp := float64(time.Now().Unix())
					suspiciousKey := "suspicious_attempts:" + clientIP
					attemptData := fmt.Sprintf("%s:%d", path, time.Now().Unix())
					if err := cache.ZAdd(ctx, suspiciousKey, timestamp, attemptData); err != nil {
						logger.Debug("Failed to track suspicious path attempt",
							zap.String("ip", clientIP),
							zap.String("path", path),
							zap.Error(err))
					}
				}()

				logger.Warn("Suspicious path blocked (cached)",
					zap.String("ip", clientIP),
					zap.String("path", path),
				)
				return c.String(http.StatusForbidden, "Access forbidden")
			}

			if handlers.IsSuspiciousPath(path) {
				// Cache the suspicious path and track with ZADD
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					// Cache the suspicious path
					if err := cache.Set(ctx, suspiciousCacheKey, []byte("blocked"), 5*time.Minute); err != nil {
						logger.Error("Failed to cache suspicious path",
							zap.String("path", path),
							zap.Error(err),
						)
					}

					// Track suspicious path attempt with ZADD
					timestamp := float64(time.Now().Unix())
					suspiciousKey := "suspicious_attempts:" + clientIP
					attemptData := fmt.Sprintf("%s:%d", path, time.Now().Unix())
					if err := cache.ZAdd(ctx, suspiciousKey, timestamp, attemptData); err != nil {
						logger.Debug("Failed to track suspicious path attempt",
							zap.String("ip", clientIP),
							zap.String("path", path),
							zap.Error(err))
					}

					// Clean up old suspicious attempts (keep last 30 days for analysis)
					thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Unix()
					if _, err := cache.ZRemRangeByScore(ctx, suspiciousKey, "0", fmt.Sprintf("%d", thirtyDaysAgo)); err != nil {
						logger.Debug("Failed to cleanup old suspicious attempts",
							zap.String("ip", clientIP),
							zap.Error(err))
					}
				}()

				logger.Warn("Suspicious path blocked",
					zap.String("ip", clientIP),
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

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodPost ||
				c.Request().Method == http.MethodPut ||
				c.Request().Method == http.MethodPatch {
				if c.Request().Header.Get("Content-Length") == "" {
					return echo.NewHTTPError(http.StatusLengthRequired, "Content-Length header required")
				}
			}
			return next(c)
		}
	})

	e.Use(middleware.BodyLimit("4M"))

	if secured {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Set("Set-Cookie",
					strings.ReplaceAll(c.Response().Header().Get("Set-Cookie"), "; Path=", "; HttpOnly; Secure; SameSite=None; Path="))
				return next(c)
			}
		})
	} else {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Set("Set-Cookie",
					strings.ReplaceAll(
						c.Response().Header().Get("Set-Cookie"),
						"; Path=",
						"; HttpOnly; SameSite=Lax; Path=",
					))
				return next(c)
			}
		})
	}

	rateLimiterConfig := RateLimiterConfig{
		RequestsPerSecond: 20,
		BurstCapacity:     100,
		WindowDuration:    1 * time.Minute,
		KeyPrefix:         "horizon_api",
	}
	rateLimiter := NewRateLimiter(cache, logger, rateLimiterConfig)

	e.Use(rateLimiter.RateLimitMiddleware(func(c echo.Context) string {
		ip := handlers.GetClientIP(c)
		userAgent := handlers.GetUserAgent(c)
		return fmt.Sprintf("%s:%s", ip, userAgent)
	}))

	// CORS middleware
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

	// Request logging middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		LogLatency:  true,
		HandleError: true,

		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("REQUEST",
					zap.String("remote_ip", c.RealIP()),
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("latency", v.Latency.String()),
					zap.String("user_agent", handlers.GetUserAgent(c)),
					zap.String("host", handlers.GetHost(c)),
					zap.Int64("bytes_in", c.Request().ContentLength),
				)
			} else {
				logger.Error("REQUEST_ERROR",
					zap.String("remote_ip", c.RealIP()),
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("latency", v.Latency.String()),
					zap.String("user_agent", handlers.GetUserAgent(c)),
					zap.String("host", handlers.GetHost(c)),
					zap.Int64("bytes_in", c.Request().ContentLength),
					zap.String("error", v.Error.Error()),
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

	// Timeout middleware
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Request timed out. Please try again later.",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			logger.Error("Request timeout occurred",
				zap.String("path", c.Path()),
				zap.String("method", c.Request().Method),
				zap.String("ip", handlers.GetClientIP(c)),
				zap.Error(err),
			)
		},

		Timeout: 1 * time.Minute,
	}))

	e.Use(SecurityHeadersMiddleware(secured))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Horizon API")
	})

	return &APIServiceImpl{
		service:    e,
		serverPort: serverPort,
		handler:    handlers.NewRouteHandler(),
		cache:      cache,
	}
}
func (h *APIServiceImpl) Client() *echo.Echo {
	return h.service
}
func (h *APIServiceImpl) RegisterWebRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
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

func (h *APIServiceImpl) Run(_ context.Context) error {
	grouped := h.handler.GroupedRoutes()
	h.service.GET("web/api/routes", func(c echo.Context) error {
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

func (h *APIServiceImpl) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}
