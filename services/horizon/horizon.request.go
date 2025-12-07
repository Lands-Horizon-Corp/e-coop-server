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

type SecurityHeaderConfig struct {
	ContentTypeNosniff    string
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreloadEnabled    bool
	ReferrerPolicy        string
	ContentSecurityPolicy string
}

type ExtendedSecurityHeaders struct {
	PermissionsPolicy             string
	ExpectCT                      string
	XPermittedCrossDomainPolicies string
	CrossOriginEmbedderPolicy     string
	CrossOriginOpenerPolicy       string
	CrossOriginResourcePolicy     string
}

type APIServiceImpl struct {
	service    *echo.Echo
	serverPort int
	handler    *handlers.RouteHandler
	cache      CacheService
	secured    bool
	logger     *zap.Logger
}

type APIService interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Client() *echo.Echo
	RegisterWebRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
	RegisterMobileRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

func getProductionSecurityConfig() SecurityHeaderConfig {
	return SecurityHeaderConfig{
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'nonce-{random}'; style-src 'self' 'nonce-{random}'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https:; media-src 'self'; object-src 'none'; frame-src 'none'; frame-ancestors 'none'; form-action 'self'; base-uri 'self'; manifest-src 'self'; worker-src 'self'; sandbox allow-scripts allow-same-origin allow-forms; require-trusted-types-for 'script'; report-uri /api/csp-violations; report-to csp-endpoint;",
	}
}

func getDevelopmentSecurityConfig() SecurityHeaderConfig {
	return SecurityHeaderConfig{
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            0,
		HSTSIncludeSubdomains: false,
		HSTSPreloadEnabled:    false,
		ReferrerPolicy:        "no-referrer-when-downgrade",
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval' http: https:; img-src 'self' data: https: http:; connect-src 'self' ws: wss: http: https:; frame-src 'self' http: https:; form-action 'self' http: https:;",
	}
}

func getExtendedSecurityHeaders() ExtendedSecurityHeaders {
	return ExtendedSecurityHeaders{
		PermissionsPolicy:             "accelerometer=(), ambient-light-sensor=(), autoplay=(), battery=(), camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), xr-spatial-tracking=(), clipboard-read=(), clipboard-write=(), gamepad=(), speaker-selection=(), vibrate=()",
		ExpectCT:                      "max-age=86400, enforce",
		XPermittedCrossDomainPolicies: "none",
		CrossOriginEmbedderPolicy:     "require-corp",
		CrossOriginOpenerPolicy:       "same-origin",
		CrossOriginResourcePolicy:     "same-origin",
	}
}

func applyExtendedSecurityHeaders(c echo.Context, headers ExtendedSecurityHeaders) {
	c.Response().Header().Set("Permissions-Policy", headers.PermissionsPolicy)
	c.Response().Header().Set("Expect-CT", headers.ExpectCT)
	c.Response().Header().Set("X-Permitted-Cross-Domain-Policies", headers.XPermittedCrossDomainPolicies)
	c.Response().Header().Set("Cross-Origin-Embedder-Policy", headers.CrossOriginEmbedderPolicy)
	c.Response().Header().Set("Cross-Origin-Opener-Policy", headers.CrossOriginOpenerPolicy)
	c.Response().Header().Set("Cross-Origin-Resource-Policy", headers.CrossOriginResourcePolicy)
	c.Response().Header().Set("Server", "")
	c.Response().Header().Set("X-Powered-By", "")
}

func SecurityHeadersMiddleware(secured bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			var securityConfig SecurityHeaderConfig
			if secured {
				securityConfig = getProductionSecurityConfig()
				extendedHeaders := getExtendedSecurityHeaders()
				applyExtendedSecurityHeaders(c, extendedHeaders)

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
			} else {
				securityConfig = getDevelopmentSecurityConfig()
			}

			c.Response().Header().Set("X-Content-Type-Options", securityConfig.ContentTypeNosniff)
			c.Response().Header().Set("X-Frame-Options", securityConfig.XFrameOptions)
			c.Response().Header().Set("Referrer-Policy", securityConfig.ReferrerPolicy)
			c.Response().Header().Set("Content-Security-Policy", securityConfig.ContentSecurityPolicy)

			return err
		}
	}
}

func NewHorizonAPIService(cache CacheService, serverPort int, secured bool) APIService {
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
				logger.Error("Firewall cache error", zap.String("ip", clientIP), zap.Error(err))
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
						logger.Debug("Failed to track blocked IP attempt", zap.String("ip", clientIP), zap.Error(err))
					}

					if err := cache.ZAdd(ctx, "blocked_ips_registry", timestamp, clientIP); err != nil {
						logger.Debug("Failed to update blocked IPs registry", zap.String("ip", clientIP), zap.Error(err))
					}

					sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()
					if _, err := cache.ZRemRangeByScore(ctx, attemptKey, "0", fmt.Sprintf("%d", sevenDaysAgo)); err != nil {
						logger.Debug("Failed to cleanup old blocked attempts", zap.String("ip", clientIP), zap.Error(err))
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

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if handlers.IsSuspiciousPath(c.Request().URL.Path) {
				clientIP := handlers.GetClientIP(c)
				host := handlers.GetHost(c)

				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()

					cacheKey := "blocked_ip:" + clientIP
					if err := cache.Set(ctx, cacheKey, []byte(host), 24*time.Hour); err != nil {
						logger.Error("Failed to ban IP", zap.String("ip", clientIP), zap.Error(err))
					}

					timestamp := float64(time.Now().Unix())
					if err := cache.ZAdd(ctx, "banned_ips_registry", timestamp, clientIP); err != nil {
						logger.Debug("Failed to update banned IPs registry", zap.String("ip", clientIP), zap.Error(err))
					}
				}()

				logger.Warn("Suspicious path detected - IP banned",
					zap.String("path", c.Request().URL.Path),
					zap.String("ip", clientIP),
					zap.String("user_agent", handlers.GetUserAgent(c)),
				)
				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}
			return next(c)
		}
	})

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodPost || c.Request().Method == http.MethodPut || c.Request().Method == http.MethodPatch {
				if c.Request().Header.Get("Content-Length") == "" {
					return echo.NewHTTPError(http.StatusLengthRequired, "Content-Length header required")
				}
			}
			return next(c)
		}
	})

	e.Use(middleware.BodyLimit("4M"))

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
		secured:    secured,
		logger:     logger,
	}
}

func (h *APIServiceImpl) Client() *echo.Echo {
	return h.service
}

func (h *APIServiceImpl) RegisterWebRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	webRoute := handlers.Route{
		Route:        "/web/" + route.Route,
		Request:      route.Request,
		Response:     route.Response,
		RequestType:  route.RequestType,
		ResponseType: route.ResponseType,
		Method:       route.Method,
		Note:         route.Note,
		Private:      route.Private,
	}

	method := strings.ToUpper(strings.TrimSpace(webRoute.Method))
	if err := h.handler.AddRoute(webRoute); err != nil {
		panic(err)
	}

	webMiddleware := append([]echo.MiddlewareFunc{h.webMiddleware()}, m...)

	switch method {
	case http.MethodGet:
		h.service.GET(webRoute.Route, callback, webMiddleware...)
	case http.MethodPost:
		h.service.POST(webRoute.Route, callback, webMiddleware...)
	case http.MethodPut:
		h.service.PUT(webRoute.Route, callback, webMiddleware...)
	case http.MethodPatch:
		h.service.PATCH(webRoute.Route, callback, webMiddleware...)
	case http.MethodDelete:
		h.service.DELETE(webRoute.Route, callback, webMiddleware...)
	}
}

func (h *APIServiceImpl) RegisterMobileRoute(route handlers.Route, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	mobileRoute := handlers.Route{
		Route:        "/mobile/" + route.Route,
		Request:      route.Request,
		Response:     route.Response,
		RequestType:  route.RequestType,
		ResponseType: route.ResponseType,
		Method:       route.Method,
		Note:         route.Note,
		Private:      route.Private,
	}

	method := strings.ToUpper(strings.TrimSpace(mobileRoute.Method))
	if err := h.handler.AddRoute(mobileRoute); err != nil {
		panic(err)
	}

	mobileMiddleware := append([]echo.MiddlewareFunc{h.mobileMiddleware()}, m...)

	switch method {
	case http.MethodGet:
		h.service.GET(mobileRoute.Route, callback, mobileMiddleware...)
	case http.MethodPost:
		h.service.POST(mobileRoute.Route, callback, mobileMiddleware...)
	case http.MethodPut:
		h.service.PUT(mobileRoute.Route, callback, mobileMiddleware...)
	case http.MethodPatch:
		h.service.PATCH(mobileRoute.Route, callback, mobileMiddleware...)
	case http.MethodDelete:
		h.service.DELETE(mobileRoute.Route, callback, mobileMiddleware...)
	}
}

func (h *APIServiceImpl) Run(_ context.Context) error {
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

func (h *APIServiceImpl) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}

func (h *APIServiceImpl) webMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
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

			if !h.secured {
				origins = append(origins,
					"http://localhost:8000",
					"http://localhost:8001",
					"http://localhost:3000",
					"http://localhost:3001",
					"http://localhost:3002",
					"http://localhost:3003",
				)
			}

			allowedHosts := make([]string, 0, len(origins))
			for _, origin := range origins {
				hostname := strings.TrimPrefix(origin, "https://")
				hostname = strings.TrimPrefix(hostname, "http://")
				allowedHosts = append(allowedHosts, hostname)
			}

			host := handlers.GetHost(c)
			if !slices.Contains(allowedHosts, host) {
				return c.String(http.StatusForbidden, "Host not allowed")
			}

			origin := c.Request().Header.Get("Origin")
			if slices.Contains(origins, origin) {
				c.Response().Header().Set("Access-Control-Allow-Origin", origin)
				c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
				c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
				c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-CSRF-Token, X-Longitude, X-Latitude, Location, X-Device-Type, X-User-Agent")
				c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")
				c.Response().Header().Set("Access-Control-Max-Age", "3600")
			}

			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}

			if h.secured {
				c.Response().Header().Set("Set-Cookie",
					strings.ReplaceAll(c.Response().Header().Get("Set-Cookie"), "; Path=", "; HttpOnly; Secure; SameSite=None; Path="))
			}

			return next(c)
		}
	}
}

func (h *APIServiceImpl) mobileMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Set-Cookie",
				strings.ReplaceAll(
					c.Response().Header().Get("Set-Cookie"),
					"; Path=",
					"; HttpOnly; SameSite=Lax; Path=",
				))
			return next(c)
		}
	}
}
