package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"

	"github.com/GlebRadaev/shlink/internal/api"

	"github.com/GlebRadaev/shlink/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Application struct {
	Ctx      context.Context
	Config   *config.Config
	Logger   *logger.Logger
	Services *service.Services
	Server   *http.Server
}

func NewApplication(ctx context.Context) *Application {
	return &Application{Ctx: ctx}
}

func (app *Application) Init() error {
	var err error
	app.Config, err = config.ParseAndLoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	app.Logger, err = logger.NewLogger("info")
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}

	repositories := repository.NewRepositoryFactory(app.Ctx, app.Config)
	app.Services = service.NewServiceFactory(app.Config, app.Logger, repositories)
	router := app.setupRoutes()

	app.Server = &http.Server{
		Addr:    app.Config.ServerAddress,
		Handler: router,
	}
	return nil
}

func (app *Application) Start() error {
	go func() {
		app.Logger.Infoln("Server started at", app.Config.ServerAddress)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatalf("Server error: %v", err)
		}
	}()
	<-app.Ctx.Done()
	return app.shutdown()
}

func (app *Application) shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		app.Logger.Errorf("Error during server shutdown: %v", err)
	} else {
		app.Logger.Info("Server shutdown successfully")
	}
	if err := app.Services.URLService.SaveData(); err != nil {
		app.Logger.Errorf("Failed to save data: %v", err)
	} else {
		app.Logger.Info("Data successfully saved before shutdown")
	}
	return nil
}

func (app *Application) setupRoutes() *chi.Mux {
	router := chi.NewRouter()
	middleware.Middleware(router)
	urlHandlers := handlers.NewURLHandlers(app.Services.URLService)
	healthHandlers := handlers.NewHealthHandlers(app.Services.HealthService)
	api.Routes(router, urlHandlers, healthHandlers)
	return router
}
