package router

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/my-tashabbus/api/internal/http/handlers"
	localmiddleware "github.com/my-tashabbus/api/internal/http/middleware"
)

type Config struct {
	ServiceName        string
	CORSAllowedOrigins []string
	Logger             *slog.Logger
}

func New(cfg Config) http.Handler {
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(localmiddleware.RequestLogger(log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", handlers.HealthHandler{ServiceName: cfg.ServiceName}.ServeHTTP)
	return r
}
