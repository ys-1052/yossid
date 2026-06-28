package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestIDMiddleware assigns a unique request ID to each request context and response header.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = uuid.New().String()
			}
			c.Set("request_id", reqID)
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)
			return next(c)
		}
	}
}

// SecureHeadersMiddleware injects standard security headers to all responses.
func SecureHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			res.Header().Set("X-Content-Type-Options", "nosniff")
			res.Header().Set("X-Frame-Options", "DENY")
			res.Header().Set("Referrer-Policy", "no-referrer")
			return next(c)
		}
	}
}

// NoCacheMiddleware sets headers preventing caching for sensitive endpoints.
func NoCacheMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			res.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			res.Header().Set("Pragma", "no-cache")
			res.Header().Set("Expires", "0")
			return next(c)
		}
	}
}

// AccessLogMiddleware logs HTTP requests without leaking sensitive parameters.
func AccessLogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Run next handlers
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			stop := time.Now()
			reqID := c.Get("request_id")

			// Custom log entry to ensure sensitive parameters (tokens, passwords) are NOT logged.
			// Path and method are logged. Raw query string, Authorization and Cookie headers are excluded.
			log.Printf("[ACCESS] time=%s request_id=%v method=%s path=%s status=%d latency=%v remote_ip=%s",
				start.Format(time.RFC3339),
				reqID,
				req.Method,
				req.URL.Path, // path only, not raw query string
				res.Status,
				stop.Sub(start),
				c.RealIP(),
			)

			return nil
		}
	}
}

// CustomHTTPErrorHandler formats responses cleanly for HTML, OIDC redirect, and JSON OAuth endpoints.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	var msg interface{} = "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
	}

	reqID := c.Get("request_id")
	log.Printf("[ERROR] request_id=%v status=%d error=%v", reqID, code, err)

	// Determine response type
	// If OIDC client request or JSON API, respond with standard OAuth Error JSON.
	// Otherwise, return standard clean JSON or minimal HTML.
	if !c.Response().Committed {
		if c.Request().Header.Get("Accept") == "application/json" || c.Response().Header().Get("Content-Type") == "application/json" {
			c.JSON(code, map[string]interface{}{
				"error":             "server_error",
				"error_description": fmt.Sprintf("%v", msg),
				"request_id":        reqID,
			})
		} else {
			c.JSON(code, map[string]interface{}{
				"error":      msg,
				"request_id": reqID,
			})
		}
	}
}
