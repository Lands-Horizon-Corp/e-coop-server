package horizon

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/go2ts/go2ts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
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

type APIImpl struct {
	service    *echo.Echo
	serverPort int
	cache      *CacheImpl
	secured    bool

	routesList     []Route
	interfacesList []APIInterfaces
}
type APIInterfaces struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type GroupedRoute struct {
	Key    string  `json:"key"`
	Routes []Route `json:"routes"`
}
type API struct {
	GroupedRoutes []GroupedRoute  `json:"grouped_routes"`
	Requests      []APIInterfaces `json:"requests"`
	Responses     []APIInterfaces `json:"responses"`
}

func getProductionSecurityConfig() SecurityHeaderConfig {
	return SecurityHeaderConfig{
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; " +
			"script-src 'self' 'nonce-{random}'; " +
			"style-src 'self' 'nonce-{random}'; " +
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
			"sandbox allow-scripts allow-same-origin allow-forms; " +
			"require-trusted-types-for 'script'; " +
			"report-uri /api/csp-violations; " +
			"report-to csp-endpoint;",
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
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval' http: https:; " +
			"img-src 'self' data: https: http:; " +
			"connect-src 'self' ws: wss: http: https:; " +
			"frame-src 'self' http: https:; " +
			"form-action 'self' http: https:;",
	}
}
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
			if secured {
				securityConfig := getProductionSecurityConfig()
				extendedHeaders := getExtendedSecurityHeaders()

				c.Response().Header().Set("X-Content-Type-Options", securityConfig.ContentTypeNosniff)
				c.Response().Header().Set("X-Frame-Options", securityConfig.XFrameOptions)
				c.Response().Header().Set("Referrer-Policy", securityConfig.ReferrerPolicy)
				c.Response().Header().Set("Content-Security-Policy", securityConfig.ContentSecurityPolicy)

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
				applyExtendedSecurityHeaders(c, extendedHeaders)
			} else {
				securityConfig := getDevelopmentSecurityConfig()
				c.Response().Header().Set("X-Content-Type-Options", securityConfig.ContentTypeNosniff)
				c.Response().Header().Set("X-Frame-Options", securityConfig.XFrameOptions)
				c.Response().Header().Set("Referrer-Policy", securityConfig.ReferrerPolicy)
				c.Response().Header().Set("Content-Security-Policy", securityConfig.ContentSecurityPolicy)
			}

			return err
		}
	}
}

func NewAPIImpl(
	cache *CacheImpl,
	serverPort int,
	secured bool,
) *APIImpl {
	return &APIImpl{
		service:        echo.New(),
		serverPort:     serverPort,
		cache:          cache,
		routesList:     []Route{},
		interfacesList: []APIInterfaces{},
		secured:        secured,
	}
}

func (h *APIImpl) Init() error {
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			if !strings.Contains(err.Error(), "sync /dev/stderr") &&
				!strings.Contains(err.Error(), "sync /dev/stdout") &&
				!strings.Contains(err.Error(), "invalid argument") {
				log.Printf("logger.Sync() error: %v\n", err)
			}
		}
	}()

	origins := []string{
		"https://ecoop-suite.netlify.app",
		"https://ecoop-suite.com",
		"https://www.ecoop-suite.com",

		"https://e-coop-member-portal-development.up.railway.app",
		"https://e-coop-member-portal-production.up.railway.app",
		"https://e-coop-member-portal-staging.up.railway.app",

		"https://development.ecoop-suite.com",
		"https://www.development.ecoop-suite.com",
		"https://staging.ecoop-suite.com",
		"https://www.staging.ecoop-suite.com",

		"https://cooperatives-development.fly.dev",
		"https://cooperatives-staging.fly.dev",
		"https://cooperatives-production.fly.dev",

		"https://cooperatives-development-production-0fc5.up.railway.app",
		"https://e-coop-server-development.up.railway.app",
		"https://e-coop-server-production.up.railway.app",
		"https://e-coop-server-staging.up.railway.app",

		"https://e-coop-client-development.up.railway.app",
		"https://e-coop-client-production.up.railway.app",
		"https://e-coop-client-staging.up.railway.app",

		"https://e-coop-member-portal-development.up.railway.app/",
		"https://e-coop-member-portal-production.up.railway.app/",
		"https://e-coop-member-portal-staging.up.railway.app/",
	}

	if !h.secured {
		origins = append(origins,
			"http://localhost:8000",
			"http://localhost:8001",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:3003",
			"http://localhost:4173",
			"http://localhost:4174",
		)
	}
	allowedHosts := make([]string, 0, len(origins))
	for _, origin := range origins {
		hostname := strings.TrimPrefix(origin, "https://")
		hostname = strings.TrimPrefix(hostname, "http://")
		allowedHosts = append(allowedHosts, hostname)
	}

	h.service.Use(middleware.Recover())
	h.service.Pre(middleware.RemoveTrailingSlash())

	if h.secured {
		h.service.Pre(middleware.HTTPSRedirect())
	}

	h.service.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			host := helpers.GetHost(c)
			if slices.Contains(allowedHosts, host) {
				return next(c)
			}
			return c.String(http.StatusForbidden, "Host not allowed")
		}
	})

	h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
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

	h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := helpers.GetClientIP(c)
			if net.ParseIP(clientIP) == nil {
				logger.Warn("Invalid IP format detected - potential attack",
					zap.String("raw_ip", clientIP),
					zap.String("user_agent", helpers.GetUserAgent(c)),
				)
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
			}

			cacheKey := "blocked_ip:" + clientIP
			hostBytes, err := h.cache.Get(c.Request().Context(), cacheKey)
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
					if err := h.cache.ZAdd(ctx, attemptKey, timestamp, fmt.Sprintf("%s:%d", c.Request().URL.Path, time.Now().Unix())); err != nil {
						logger.Debug("Failed to track blocked IP attempt", zap.String("ip", clientIP), zap.Error(err))
					}

					if err := h.cache.ZAdd(ctx, "blocked_ips_registry", timestamp, clientIP); err != nil {
						logger.Debug("Failed to update blocked IPs registry", zap.String("ip", clientIP), zap.Error(err))
					}
					sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()
					if _, err := h.cache.ZRemRangeByScore(ctx, attemptKey, "0", fmt.Sprintf("%d", sevenDaysAgo)); err != nil {
						logger.Debug("Failed to cleanup old blocked attempts", zap.String("ip", clientIP), zap.Error(err))
					}
				}()
				logger.Warn("Blocked IP access attempt", zap.String("ip", clientIP), zap.String("blocked_host", blockedHost), zap.String("path", c.Request().URL.Path))
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Access denied",
					"code":  "IP_BLOCKED",
				})
			}
			return next(c)
		}
	})

	h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := helpers.GetPath(c)
			clientIP := helpers.GetClientIP(c)
			suspiciousCacheKey := "suspicious_path:" + path
			cachedResult, err := h.cache.Get(c.Request().Context(), suspiciousCacheKey)
			if err == nil && cachedResult != nil {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					timestamp := float64(time.Now().Unix())
					suspiciousKey := "suspicious_attempts:" + clientIP
					attemptData := fmt.Sprintf("%s:%d", path, time.Now().Unix())
					if err := h.cache.ZAdd(ctx, suspiciousKey, timestamp, attemptData); err != nil {
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
			if helpers.IsSuspicious(path) {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := h.cache.Set(ctx, suspiciousCacheKey, []byte("blocked"), 5*time.Minute); err != nil {
						logger.Error("Failed to cache suspicious path",
							zap.String("path", path),
							zap.Error(err),
						)
					}
					timestamp := float64(time.Now().Unix())
					suspiciousKey := "suspicious_attempts:" + clientIP
					attemptData := fmt.Sprintf("%s:%d", path, time.Now().Unix())
					if err := h.cache.ZAdd(ctx, suspiciousKey, timestamp, attemptData); err != nil {
						logger.Debug("Failed to track suspicious path attempt",
							zap.String("ip", clientIP),
							zap.String("path", path),
							zap.Error(err))
					}
					thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Unix()
					if _, err := h.cache.ZRemRangeByScore(ctx, suspiciousKey, "0", fmt.Sprintf("%d", thirtyDaysAgo)); err != nil {
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

	h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
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

	h.service.Use(middleware.BodyLimit("4M"))
	if h.secured {
		h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Set("Set-Cookie",
					strings.ReplaceAll(c.Response().Header().Get("Set-Cookie"), "; Path=", "; HttpOnly; Secure; SameSite=None; Path="))
				return next(c)
			}
		})
	} else {
		h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
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
	h.service.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(10), Burst: 30, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}))

	h.service.Use(middleware.CORSWithConfig(middleware.CORSConfig{
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
			"X-Organization-id",
		},

		ExposeHeaders: []string{
			echo.HeaderContentLength,
			echo.HeaderContentType,
			echo.HeaderAuthorization,
		},

		AllowCredentials: true,
		MaxAge:           3600,
	}))

	h.service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
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

	h.service.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
					zap.String("user_agent", helpers.GetUserAgent(c)),
					zap.String("host", helpers.GetHost(c)),
					zap.Int64("bytes_in", c.Request().ContentLength),
				)
			} else {
				logger.Error("REQUEST_ERROR",
					zap.String("remote_ip", c.RealIP()),
					zap.String("method", v.Method),
					zap.String("uri", v.URI),
					zap.Int("status", v.Status),
					zap.String("latency", v.Latency.String()),
					zap.String("user_agent", helpers.GetUserAgent(c)),
					zap.String("host", helpers.GetHost(c)),
					zap.Int64("bytes_in", c.Request().ContentLength),
					zap.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	h.service.Use(middleware.GzipWithConfig(middleware.GzipConfig{
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

	h.service.Use(SecurityHeadersMiddleware(h.secured))
	h.service.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Horizon API")
	})
	return nil
}

func (h *APIImpl) Run() error {

	grouped := h.GroupedRoutes()
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

func (h *APIImpl) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}

func (h *APIImpl) RegisterWebRoute(
	route Route,
	callback func(c echo.Context) error,
	mid ...echo.MiddlewareFunc,
) {
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	route.Route = "/web/" + strings.TrimPrefix(route.Route, "/")

	switch method {
	case http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodDelete:
	default:
		panic(eris.Errorf(
			"unsupported HTTP method: %s for route: %s",
			method,
			route.Route,
		))
	}

	for _, existing := range h.routesList {
		if strings.EqualFold(existing.Route, route.Route) &&
			strings.EqualFold(existing.Method, method) {
			panic(eris.Errorf(
				"route already registered: %s %s",
				method,
				route.Route,
			))
		}
	}

	if route.Private {
		return
	}

	request := go2ts.Convert(route.RequestType)
	response := go2ts.Convert(route.ResponseType)

	if route.RequestType != nil {
		h.interfacesList = append(h.interfacesList, APIInterfaces{
			Key:   helpers.ExtractInterfaceName(route.RequestType),
			Value: request,
		})
	}

	if route.ResponseType != nil {
		h.interfacesList = append(h.interfacesList, APIInterfaces{
			Key:   helpers.ExtractInterfaceName(route.ResponseType),
			Value: response,
		})
	}

	h.routesList = append(h.routesList, Route{
		Route:    route.Route,
		Method:   method,
		Request:  request,
		Response: response,
		Note:     route.Note,
	})

	switch method {
	case http.MethodGet:
		h.service.GET(route.Route, callback, mid...)
	case http.MethodPost:
		h.service.POST(route.Route, callback, mid...)
	case http.MethodPut:
		h.service.PUT(route.Route, callback, mid...)
	case http.MethodPatch:
		h.service.PATCH(route.Route, callback, mid...)
	case http.MethodDelete:
		h.service.DELETE(route.Route, callback, mid...)
	}
}

func (h *APIImpl) GroupedRoutes() API {
	ignoredPrefixes := []string{"api", "v1", "v2", "web", "mobile"}
	groupMap := make(map[string][]Route)
	requestMap := make(map[string]APIInterfaces)
	responseMap := make(map[string]APIInterfaces)
	isIgnored := func(part string) bool {
		return slices.Contains(ignoredPrefixes, part)
	}
	for _, r := range h.routesList {
		if r.Private {
			continue
		}
		parts := strings.Split(strings.Trim(r.Route, "/"), "/")
		base := "/"
		for _, part := range parts {
			if part != "" && !isIgnored(part) {
				base = part
				break
			}
		}
		groupMap[base] = append(groupMap[base], r)
		if r.Request != "" {
			key := helpers.ExtractInterfaceName(r.RequestType)
			if _, exists := requestMap[key]; !exists {
				requestMap[key] = APIInterfaces{
					Key:   key,
					Value: r.Request,
				}
			}
		}
		if r.Response != "" {
			key := helpers.ExtractInterfaceName(r.ResponseType)
			if _, exists := responseMap[key]; !exists {
				responseMap[key] = APIInterfaces{
					Key:   key,
					Value: r.Response,
				}
			}
		}
	}
	groupedRoutes := make([]GroupedRoute, 0, len(groupMap))
	for k, routes := range groupMap {
		groupedRoutes = append(groupedRoutes, GroupedRoute{
			Key:    k,
			Routes: routes,
		})
	}
	requests := make([]APIInterfaces, 0, len(requestMap))
	for _, v := range requestMap {
		requests = append(requests, v)
	}
	responses := make([]APIInterfaces, 0, len(responseMap))
	for _, v := range responseMap {
		responses = append(responses, v)
	}
	return API{
		GroupedRoutes: groupedRoutes,
		Requests:      requests,
		Responses:     responses,
	}
}
