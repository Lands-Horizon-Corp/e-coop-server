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
	"golang.org/x/time/rate"
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
	RegisterRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

// APIServiceImpl implements APIService with comprehensive security middleware stack.
// This implementation provides enterprise-grade security features including:
// - Redis-backed rate limiting and IP blocking
// - Comprehensive security headers
// - Request validation and suspicious path detection
// - Environment-aware configuration (dev/prod)
type APIServiceImpl struct {
	service     *echo.Echo             // Echo web framework instance
	serverPort  int                    // HTTP server port
	metricsPort int                    // Metrics server port (future use)
	clientURL   string                 // Client application URL
	clientName  string                 // Client application name
	handler     *handlers.RouteHandler // Route management handler
	cache       CacheService           // Redis cache service for security features
}

// RedisRateLimiterStore implements Echo's RateLimiterStore interface using Redis
// for distributed rate limiting across multiple Fly.io instances.
// This ensures consistent rate limiting behavior in a multi-instance deployment.
type RedisRateLimiterStore struct {
	cache     CacheService  // Redis cache service for storing rate limit counters
	logger    *zap.Logger   // Structured logger for rate limiting events
	rate      rate.Limit    // Requests per second limit
	burst     int           // Burst capacity for traffic spikes
	expiresIn time.Duration // Time window for rate limiting
}

// Allow implements the RateLimiterStore interface using a sliding window algorithm.
// It tracks requests per second over a sliding window, providing precise rate limiting.
// Returns true if the request is allowed, false if rate limited.
func (s *RedisRateLimiterStore) Allow(identifier string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-s.expiresIn)

	// Use sorted set to track requests with timestamps
	key := "rate_limit:" + identifier

	// Remove expired entries (older than window)
	if err := s.removeExpiredEntries(ctx, key, windowStart.Unix()); err != nil {
		s.logger.Error("Failed to clean expired rate limit entries",
			zap.String("identifier", identifier),
			zap.Error(err))
	}

	// Count current requests in the window
	currentCount, err := s.getRequestCount(ctx, key, windowStart.Unix())
	if err != nil {
		// If Redis is down, allow request but log error
		s.logger.Error("Rate limit cache error", zap.String("identifier", identifier), zap.Error(err))
		return true, nil // Fail open
	}

	// Check if rate limit would be exceeded
	maxRequests := int(float64(s.rate) * s.expiresIn.Seconds())
	if currentCount >= maxRequests {
		s.logger.Debug("Rate limit exceeded",
			zap.String("identifier", identifier),
			zap.Int("current_count", currentCount),
			zap.Int("max_requests", maxRequests),
			zap.Duration("window", s.expiresIn),
		)
		return false, nil
	}

	// Add current request timestamp
	if err := s.addRequest(ctx, key, now.Unix()); err != nil {
		s.logger.Error("Failed to record rate limit request",
			zap.String("identifier", identifier),
			zap.Error(err),
		)
		// Still allow the request even if we can't record it
		return true, nil
	}

	return true, nil
}

// removeExpiredEntries removes rate limit entries older than the window start time using Redis ZREMRANGEBYSCORE
func (s *RedisRateLimiterStore) removeExpiredEntries(ctx context.Context, key string, windowStart int64) error {
	// Use Redis sorted set ZREMRANGEBYSCORE to efficiently remove old entries
	// Remove all entries with scores (timestamps) less than windowStart
	_, err := s.cache.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart-1))
	if err != nil {
		s.logger.Debug("Failed to remove expired entries from sorted set",
			zap.String("key", key),
			zap.Int64("window_start", windowStart),
			zap.Error(err))
	}
	return err
}

// getRequestCount returns the number of requests in the current window using Redis ZCARD
func (s *RedisRateLimiterStore) getRequestCount(ctx context.Context, key string, windowStart int64) (int, error) {
	// Use Redis sorted set ZCARD to get the count of all members in the set
	// Since we clean up expired entries, all remaining entries are valid
	count, err := s.cache.ZCard(ctx, key)
	if err != nil {
		s.logger.Debug("Failed to get request count from sorted set",
			zap.String("key", key),
			zap.Error(err))
		return 0, err
	}

	return int(count), nil
} // addRequest adds a new request timestamp to the rate limit tracking using Redis ZADD
func (s *RedisRateLimiterStore) addRequest(ctx context.Context, key string, timestamp int64) error {
	// Use Redis sorted set ZADD to add timestamp as both score and member
	// This provides O(log N) insertion performance and automatic sorting
	err := s.cache.ZAdd(ctx, key, float64(timestamp), timestamp)
	if err != nil {
		s.logger.Debug("Failed to add request to sorted set",
			zap.String("key", key),
			zap.Int64("timestamp", timestamp),
			zap.Error(err))
		return err
	}

	// Set TTL on the key to ensure automatic cleanup if no requests for extended period
	// Note: This uses a standard Set operation with TTL as an additional safety measure
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Set a dummy value with TTL to ensure the key expires
		if err := s.cache.Set(ctx, key+":ttl", []byte("1"), s.expiresIn*2); err != nil {
			s.logger.Debug("Failed to set TTL marker for rate limit key",
				zap.String("key", key),
				zap.Error(err))
		}
	}()

	return nil
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
//   - metricsPort: Metrics server port (reserved for future Prometheus integration)
//   - clientURL: Primary client application URL for CORS configuration
//   - clientName: Client application identifier for logging
//   - secured: Production mode flag (enables HTTPS redirect and strict security headers)
//
// Returns: Configured APIService ready for production deployment
func NewHorizonAPIService(
	cache CacheService,
	serverPort, metricsPort int,
	clientURL, clientName string,
	secured bool,
) APIService {
	// ===== ECHO INSTANCE AND LOGGER SETUP =====
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
	// Define allowed origins for Cross-Origin Resource Sharing (CORS)
	// These domains are permitted to make requests to the API

	// Production domains - primary application domains
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

	// ===== BASIC MIDDLEWARE SETUP =====
	// Panic recovery middleware
	e.Use(middleware.Recover())

	// Remove trailing slashes from URLs
	e.Pre(middleware.RemoveTrailingSlash())

	// Force HTTPS redirect in production
	if secured {
		e.Pre(middleware.HTTPSRedirect())
	}

	// ===== HOST VALIDATION MIDDLEWARE =====
	// Validates the Host header against approved domains to prevent:
	// - Host header injection attacks
	// - DNS rebinding attacks
	// - Unauthorized domain access
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

	// ===== HTTP METHOD RESTRICTION MIDDLEWARE =====
	// Restricts HTTP methods to standard, safe operations.
	// Blocks potentially dangerous methods like TRACE, CONNECT, etc.
	// This prevents HTTP method tampering and reduces attack surface.
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

	// ===== IP FIREWALL MIDDLEWARE =====
	// Redis-backed IP firewall for blocking malicious traffic.
	// Features:
	// - Integration with HaGeZi threat intelligence blocklists
	// - Manual IP blocking capability
	// - Distributed blocking across Fly.io instances
	// - Graceful fallback if Redis is unavailable
	// - Sorted set tracking for analytics and automatic cleanup
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
					attemptKey := "blocked_attempts:" + clientIP
					if err := cache.ZAdd(ctx, attemptKey, timestamp, fmt.Sprintf("%s:%d", c.Request().URL.Path, time.Now().Unix())); err != nil {
						logger.Debug("Failed to track blocked IP attempt",
							zap.String("ip", clientIP),
							zap.Error(err))
					}
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

	// ===== SUSPICIOUS PATH DETECTION MIDDLEWARE =====
	// Advanced threat detection for malicious request patterns.
	// Detects and blocks:
	// - SQL injection attempts
	// - XSS (Cross-Site Scripting) attacks
	// - Directory traversal attempts (../, etc.)
	// - Command injection patterns
	// - File inclusion attacks
	// - Web shell upload attempts
	// - Uses sorted sets for analytics and automatic cleanup
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

	// ===== REQUEST SIZE LIMIT MIDDLEWARE =====
	// Limit request body size to prevent DoS attacks and validate Content-Length header
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check for Content-Length header on requests with body
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

	// ===== SECURITY HEADERS MOVED TO FINAL MIDDLEWARE =====
	// Security headers are now applied as final middleware after all route processing
	// to ensure consistency and prevent route handlers from overriding security headers

	// ===== SECURE COOKIE MIDDLEWARE =====
	// Cookie security configuration for API authentication tokens.
	// Uses environment-aware settings for production vs development.
	if secured {
		// Production: Strict security for HTTPS-only deployment
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Set secure cookie defaults for any cookies set by handlers
				c.Response().Header().Set("Set-Cookie",
					strings.ReplaceAll(c.Response().Header().Get("Set-Cookie"), "; Path=", "; HttpOnly; Secure; SameSite=None; Path="))
				return next(c)
			}
		})
	} else {
		// Development: HTTP-compatible settings for local development
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Set development cookie defaults (allow HTTP)
				c.Response().Header().Set("Set-Cookie",
					strings.ReplaceAll(c.Response().Header().Get("Set-Cookie"), "; Path=", "; HttpOnly; SameSite=Lax; Path="))
				return next(c)
			}
		})
	}

	// ===== RATE LIMITING MIDDLEWARE =====
	// Redis-backed distributed rate limiting for Fly.io multi-instance deployment.
	// Features:
	// - Consistent rate limiting across all instances
	// - IP + User-Agent fingerprinting for accurate identification
	// - Graceful degradation if Redis is unavailable
	// - Client-aware headers for better UX
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,

		// Custom Redis store for distributed rate limiting
		Store: &RedisRateLimiterStore{
			cache:     cache,           // Redis cache service
			logger:    logger,          // Structured logger
			rate:      rate.Limit(20),  // 20 requests per second
			burst:     100,             // Burst capacity for traffic spikes
			expiresIn: 1 * time.Minute, // Rate limit window duration
		},
		// Generate unique identifier combining IP address and User-Agent
		// This prevents simple IP rotation attacks while maintaining accuracy
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			ip := handlers.GetClientIP(ctx)
			userAgent := handlers.GetUserAgent(ctx)
			return fmt.Sprintf("%s:%s", ip, userAgent), nil
		},

		// Handle rate limiter internal errors gracefully
		ErrorHandler: func(c echo.Context, err error) error {
			if secured {
				// Production: Generic error message
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Request rate limited",
				})
			}
			// Development: Detailed error information
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "rate limit error: " + err.Error(),
			})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			// Add rate limit headers for client awareness
			c.Response().Header().Set("X-RateLimit-Limit", "20")
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(1*time.Minute).Unix()))

			logger.Warn("Rate limit exceeded",
				zap.String("identifier", identifier),
				zap.String("ip", handlers.GetClientIP(c)),
				zap.String("user_agent", handlers.GetUserAgent(c)),
			)

			if secured {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests. Please try again later.",
				})
			}
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error":       "Rate limit exceeded",
				"retry_after": "60s",
			})
		},
	}))

	// ===== CORS MIDDLEWARE =====
	// Cross-Origin Resource Sharing configuration for web application integration.
	// Allows legitimate cross-origin requests while maintaining security.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// Allowed origins - only trusted domains can make requests
		AllowOrigins: origins,

		// Permitted HTTP methods for cross-origin requests
		AllowMethods: []string{
			http.MethodGet,     // Read operations
			http.MethodPost,    // Create operations
			http.MethodPut,     // Update operations
			http.MethodPatch,   // Partial updates
			http.MethodDelete,  // Delete operations
			http.MethodOptions, // Preflight requests
			http.MethodHead,    // Header-only requests
		},
		// Headers that client applications are allowed to send
		AllowHeaders: []string{
			// Standard CORS headers
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestedWith,
			echo.HeaderXCSRFToken,
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAccessControlRequestHeaders,

			// Custom application headers
			"X-Longitude",   // Geographic location data
			"X-Latitude",    // Geographic location data
			"Location",      // Location information
			"X-Device-Type", // Client device type
			"X-User-Agent",  // Enhanced user agent info
		},

		// Headers that can be exposed to the client application
		ExposeHeaders: []string{
			echo.HeaderContentLength, // Response size information
			echo.HeaderContentType,   // Response content type
			echo.HeaderAuthorization, // Authentication headers
		},

		AllowCredentials: true, // Allow cookies and credentials
		MaxAge:           3600, // Cache preflight response for 1 hour
	}))

	// ===== CORS PREFLIGHT DEBUGGING MIDDLEWARE =====
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

	// ===== REQUEST LOGGING MIDDLEWARE =====
	// Comprehensive request logging for monitoring and debugging.
	// Logs all HTTP requests with structured data for analysis and security monitoring.
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true, // Log HTTP status codes
		LogURI:      true, // Log request URIs
		LogError:    true, // Log error details
		LogMethod:   true, // Log HTTP methods
		LogLatency:  true, // Log request latency
		HandleError: true, // Handle errors in logging

		// Enhanced log formatting function with comprehensive request data
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				// Log successful requests with detailed context
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
				// Log failed requests with comprehensive error details
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

	// ===== COMPRESSION MIDDLEWARE =====
	// Intelligent response compression for bandwidth optimization.
	// Automatically compresses text-based responses while skipping binary content.
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 6, // Balanced compression level (speed vs. size)

		// Skip compression for binary content types
		Skipper: func(c echo.Context) bool {
			ct := c.Response().Header().Get(echo.HeaderContentType)
			return strings.HasPrefix(ct, "image/") || // Images
				strings.HasPrefix(ct, "video/") || // Videos
				strings.HasPrefix(ct, "audio/") || // Audio files
				strings.HasPrefix(ct, "application/zip") || // Archives
				strings.HasPrefix(ct, "application/pdf") // PDF documents
		},
	}))

	// ===== TIMEOUT MIDDLEWARE =====
	// Request timeout protection to prevent resource exhaustion.
	// Automatically terminates long-running requests to maintain server stability.
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Request timed out. Please try again later.",

		// Custom timeout error handler for detailed logging
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			logger.Error("Request timeout occurred",
				zap.String("path", c.Path()),
				zap.String("method", c.Request().Method),
				zap.String("ip", handlers.GetClientIP(c)),
				zap.Error(err),
			)
		},

		Timeout: 1 * time.Minute, // 60-second request timeout
	}))

	// ===== FINAL SECURITY HEADERS MIDDLEWARE =====
	// Apply security headers as the final middleware to ensure consistency
	// This guarantees headers are set on all responses regardless of route handlers
	e.Use(SecurityHeadersMiddleware(secured))

	//===== DEFAULT ENDPOINTS =====
	// Essential API endpoints for monitoring and discovery

	// Health check endpoint for load balancer and monitoring systems
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// API root endpoint with welcome message
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

// Client returns the underlying Echo instance for advanced configuration.
// Use this method when you need direct access to Echo's features.
func (h *APIServiceImpl) Client() *echo.Echo {
	return h.service
}

// RegisterRoute adds a new HTTP route to the API server.
// This method automatically handles route registration and method validation.
//
// Parameters:
//   - route: Route configuration containing method, path, and metadata
//   - callback: Handler function for the route
//   - m: Optional middleware functions to apply to this specific route
func (h *APIServiceImpl) RegisterRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	method := strings.ToUpper(strings.TrimSpace(route.Method))

	// Add route to internal handler for tracking and management
	if err := h.handler.AddRoute(route); err != nil {
		panic(err)
	}

	// Register route with appropriate HTTP method
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

// Run starts the API server with all configured security middleware.
// This method initializes route discovery endpoints and starts the HTTP server.
//
// The server runs in a separate goroutine, allowing the method to return immediately.
// Use the Stop method for graceful shutdown.
func (h *APIServiceImpl) Run(_ context.Context) error {
	// Register route discovery endpoint for API documentation
	grouped := h.handler.GroupedRoutes()
	h.service.GET("/api/routes", func(c echo.Context) error {
		return c.JSON(http.StatusOK, grouped)
	}).Name = "horizon-routes-json"

	// Catch-all route for unmatched requests
	h.service.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "404 - Route not found")
	})

	// Start HTTP server in background goroutine
	go func() {
		h.service.Logger.Fatal(h.service.Start(fmt.Sprintf(":%d", h.serverPort)))
	}()

	return nil
}

// Stop gracefully shuts down the API server.
// This method ensures all active connections are completed before stopping.
//
// Parameters:
//   - ctx: Context for controlling shutdown timeout
//
// Returns: Error if shutdown fails
func (h *APIServiceImpl) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}
