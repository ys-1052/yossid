package app

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ys-1052/yossid/backend/internal/config"
	"github.com/ys-1052/yossid/backend/internal/handler"
	"github.com/ys-1052/yossid/backend/internal/mail"
	"github.com/ys-1052/yossid/backend/internal/repository"
	"github.com/ys-1052/yossid/backend/internal/service"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres"
)

type App struct {
	Echo   *echo.Echo
	Config *config.Config
	DB     *postgres.DB
}

func NewApp(ctx context.Context) (*App, error) {
	e := echo.New()

	// Enable standard recover middleware
	e.Use(middleware.Recover())

	// Set custom HTTP error handler
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	// Load configuration
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize Database connection pool
	pgDB, err := postgres.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize Repositories
	userRepo := repository.NewUserRepository(pgDB)
	registerRepo := repository.NewRegistrationRepository(pgDB)
	otpRepo := repository.NewOTPRepository(pgDB)
	sessionRepo := repository.NewSessionRepository(pgDB)
	auditRepo := repository.NewAuditRepository(pgDB)

	// Initialize Mailer
	mailer, err := mail.NewMailer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Initialize Services
	regService := service.NewRegistrationService(cfg, userRepo, registerRepo, mailer)
	loginService := service.NewLoginService(cfg, userRepo, otpRepo, sessionRepo, auditRepo, mailer)

	// Initialize Handlers
	healthHandler := handler.NewHealthHandler()
	registerHandler := handler.NewRegisterHandler(regService)
	loginHandler := handler.NewLoginHandler(loginService)

	handlers := &ServerHandlers{
		Health:   healthHandler,
		Register: registerHandler,
		Login:    loginHandler,
	}

	// Register routes
	RegisterRoutes(e, handlers)

	return &App{
		Echo:   e,
		Config: cfg,
		DB:     pgDB,
	}, nil
}

// Start runs the Echo application as an HTTP server.
func (a *App) Start() error {
	addr := ":" + a.Config.Port
	return a.Echo.Start(addr)
}

// Handler returns the http.Handler for AWS Lambda proxy integration.
func (a *App) Handler() http.Handler {
	return a.Echo
}
