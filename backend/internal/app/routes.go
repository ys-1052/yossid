package app

import (
	"github.com/labstack/echo/v4"
	"github.com/ys-1052/yossid/backend/internal/handler"
)

type ServerHandlers struct {
	Health   *handler.HealthHandler
	Register *handler.RegisterHandler
	Login    *handler.LoginHandler
	// Placeholders for other handlers:
	// Authorize  *handler.AuthorizeHandler
	// Token      *handler.TokenHandler
	// Userinfo   *handler.UserinfoHandler
	// Discovery  *handler.DiscoveryHandler
	// Jwks       *handler.JwksHandler
}

func RegisterRoutes(e *echo.Echo, handlers *ServerHandlers) {
	// Global Middlewares
	e.Use(RequestIDMiddleware())
	e.Use(SecureHeadersMiddleware())
	e.Use(AccessLogMiddleware())

	// Health Check
	e.GET("/healthz", handlers.Health.Healthz)

	// Auth & OIDC endpoints (No-cache required)
	authGroup := e.Group("")
	authGroup.Use(NoCacheMiddleware())

	// Registration routes
	authGroup.POST("/register", handlers.Register.PostRegister)
	authGroup.GET("/email/verify", handlers.Register.VerifyEmail)

	// Login / MFA routes
	authGroup.POST("/login", handlers.Login.PostLogin)
	authGroup.POST("/mfa/email/verify", handlers.Login.PostVerifyMFA)
	authGroup.POST("/logout", handlers.Login.PostLogout)

	// TODO: Register actual handlers here
	// authGroup.GET("/authorize", handlers.Authorize.GetAuthorize)
	// authGroup.POST("/token", handlers.Token.PostToken)
	// authGroup.GET("/userinfo", handlers.Userinfo.GetUserInfo)

	// Discovery and JWKS endpoints (Cacheable)
	// e.GET("/.well-known/openid-configuration", handlers.Discovery.GetDiscovery)
	// e.GET("/jwks.json", handlers.Jwks.GetJWKS)
}
