package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/db"
	apirouter "github.com/my-tashabbus/api/internal/http/router"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type App struct {
	cfg         config.Config
	log         *slog.Logger
	authService *auth.Service
	userService *users.Service
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	userStore := users.Store(users.NewMemoryStore())
	if cfg.DatabaseURL != "" {
		pool, err := db.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("open database: %w", err)
		}
		userStore = users.NewPgxStore(pool)
	}
	userService := users.NewService(userStore)
	authService, err := auth.NewService(cfg, userService)
	if err != nil {
		return nil, fmt.Errorf("create auth service: %w", err)
	}
	return &App{cfg: cfg, log: log, authService: authService, userService: userService}, nil
}

func (a *App) Handler() http.Handler {
	return apirouter.New(apirouter.Config{
		ServiceName:        "my-tashabbus-api",
		CORSAllowedOrigins: a.cfg.CORSAllowedOrigins,
		Logger:             a.log,
		AuthService:        a.authService,
		UserService:        a.userService,
	})
}
