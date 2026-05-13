package app

import (
	"log/slog"
	"net/http"

	"github.com/my-tashabbus/api/internal/config"
	apirouter "github.com/my-tashabbus/api/internal/http/router"
)

type App struct {
	cfg config.Config
	log *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) *App {
	return &App{cfg: cfg, log: log}
}

func (a *App) Handler() http.Handler {
	return apirouter.New(apirouter.Config{
		ServiceName:        "my-tashabbus-api",
		CORSAllowedOrigins: a.cfg.CORSAllowedOrigins,
		Logger:             a.log,
	})
}
