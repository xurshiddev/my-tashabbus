package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/db"
	apirouter "github.com/my-tashabbus/api/internal/http/router"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/households"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/responsibles"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type App struct {
	cfg                config.Config
	log                *slog.Logger
	authService        *auth.Service
	userService        *users.Service
	mfyService         *mfys.Service
	streetService      *streets.Service
	householdService   *households.Service
	responsibleService *responsibles.Service
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	var pool *pgxpool.Pool
	userStore := users.Store(users.NewMemoryStore())
	if cfg.DatabaseURL != "" {
		var err error
		pool, err = db.Open(ctx, cfg.DatabaseURL)
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
	mfyStore := mfys.Store(mfys.NewMemoryStore())
	streetStore := streets.Store(streets.NewMemoryStore())
	householdStore := households.Store(households.NewMemoryStore())
	responsibleStore := responsibles.Store(responsibles.NewMemoryStore())
	if pool != nil {
		mfyStore = mfys.NewPgxStore(pool)
		streetStore = streets.NewPgxStore(pool)
		householdStore = households.NewPgxStore(pool)
		responsibleStore = responsibles.NewPgxStore(pool)
	}
	mfyService := mfys.NewService(mfyStore, userService)
	streetService := streets.NewService(streetStore, mfyService, userService)
	householdService := households.NewService(householdStore, streetService)
	responsibleService := responsibles.NewService(responsibleStore, streetService, userService, householdService)
	return &App{cfg: cfg, log: log, authService: authService, userService: userService, mfyService: mfyService, streetService: streetService, householdService: householdService, responsibleService: responsibleService}, nil
}

func (a *App) Handler() http.Handler {
	return apirouter.New(apirouter.Config{
		ServiceName:        "my-tashabbus-api",
		CORSAllowedOrigins: a.cfg.CORSAllowedOrigins,
		Logger:             a.log,
		AuthService:        a.authService,
		UserService:        a.userService,
		MFYService:         a.mfyService,
		StreetService:      a.streetService,
		HouseholdService:   a.householdService,
		ResponsibleService: a.responsibleService,
	})
}
