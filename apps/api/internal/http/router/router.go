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
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/households"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/responsibles"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Config struct {
	ServiceName        string
	CORSAllowedOrigins []string
	Logger             *slog.Logger
	AuthService        *auth.Service
	UserService        *users.Service
	MFYService         *mfys.Service
	StreetService      *streets.Service
	HouseholdService   *households.Service
	ResponsibleService *responsibles.Service
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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", localmiddleware.TelegramInitDataHeader, "ngrok-skip-browser-warning"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", handlers.HealthHandler{ServiceName: cfg.ServiceName}.ServeHTTP)
	if cfg.AuthService != nil && cfg.UserService != nil {
		authHandler := auth.NewHandler(cfg.AuthService, log)
		userHandler := users.NewHandler(cfg.UserService)
		mfyHandler := mfys.NewHandler(cfg.MFYService)
		streetHandler := streets.NewHandler(cfg.StreetService)
		householdHandler := households.NewHandler(cfg.HouseholdService, log)
		responsibleHandler := responsibles.NewHandler(cfg.ResponsibleService, log)
		requireAuth := localmiddleware.RequireAuthWithDiagnostics(
			cfg.AuthService.TokenManager(),
			cfg.UserService,
			log,
			cfg.AuthService.AppEnv() != "production",
		)
		requireTelegramMiniApp := localmiddleware.RequireTelegramInitData(
			cfg.AuthService,
			log,
			cfg.AuthService.AppEnv() != "production",
		)
		requireSuperAdmin := localmiddleware.RequireRole(users.RoleSuperAdmin)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/dev-login", authHandler.DevLogin)
			r.Post("/telegram", authHandler.TelegramLogin)
			r.With(requireAuth).Get("/me", authHandler.Me)
		})

		r.Route("/users", func(r chi.Router) {
			r.With(requireAuth).Get("/me", authHandler.Me)
			r.Group(func(r chi.Router) {
				r.Use(requireAuth)
				r.Use(requireSuperAdmin)
				r.Get("/", userHandler.List)
				r.Post("/", userHandler.Create)
				r.Get("/{id}", userHandler.Get)
				r.Patch("/{id}", userHandler.Update)
				r.Patch("/{id}/telegram", userHandler.SetTelegramIdentity)
				r.Post("/{id}/deactivate", userHandler.Deactivate)
				r.Post("/{id}/activate", userHandler.Activate)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(requireAuth)
			r.Get("/mfys", mfyHandler.List)
			r.Post("/mfys", mfyHandler.Create)
			r.Get("/mfys/{id}", mfyHandler.Get)
			r.Patch("/mfys/{id}", mfyHandler.Update)
			r.Post("/mfys/{id}/assign-chairman", mfyHandler.AssignChairman)
			r.Get("/mfys/{mfyID}/streets", streetHandler.ListByMFY)
			r.Post("/mfys/{mfyID}/streets", streetHandler.Create)
			r.Get("/streets/{id}", streetHandler.Get)
			r.Patch("/streets/{id}", streetHandler.Update)
			r.Post("/streets/{id}/assign-leader", streetHandler.AssignLeader)
			r.Get("/streets/{id}/leader", streetHandler.GetLeader)
			r.Get("/my/streets", streetHandler.MyStreets)
			r.Post("/streets/{streetID}/households", householdHandler.Create)
			r.Get("/streets/{streetID}/households", householdHandler.ListByStreet)
			r.Get("/households/{id}", householdHandler.Get)
			r.Patch("/households/{id}", householdHandler.Update)
			r.Get("/my/households", householdHandler.MyHouseholds)
			r.Get("/households/{id}/logs", householdHandler.Logs)
			r.Post("/streets/{streetID}/responsibles", responsibleHandler.Create)
			r.Get("/streets/{streetID}/responsibles", responsibleHandler.ListByStreet)
			r.Post("/responsible-assignments/{id}/deactivate", responsibleHandler.Deactivate)
		})

		r.Route("/miniapp", func(r chi.Router) {
			r.Use(requireTelegramMiniApp)
			r.Get("/me", authHandler.MiniAppMe)
			r.Get("/my/streets", streetHandler.MyStreets)
			r.Get("/my/households", householdHandler.MyHouseholds)
			r.Get("/streets/{streetID}/households", householdHandler.ListByStreet)
			r.Post("/streets/{streetID}/households", householdHandler.Create)
			r.Get("/streets/{streetID}/responsibles", responsibleHandler.ListByStreet)
			r.Post("/streets/{streetID}/responsibles", responsibleHandler.Create)
			r.Patch("/households/{id}", householdHandler.Update)
			r.Get("/households/{id}/logs", householdHandler.Logs)
			r.Post("/responsible-assignments/{id}/deactivate", responsibleHandler.Deactivate)
		})
	}
	return r
}
