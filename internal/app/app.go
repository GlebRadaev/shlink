package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/GlebRadaev/shlink/internal/api"
	"github.com/GlebRadaev/shlink/internal/api/handlers"
	"github.com/GlebRadaev/shlink/internal/config"
	"github.com/GlebRadaev/shlink/internal/logger"
	"github.com/GlebRadaev/shlink/internal/middleware"
	"github.com/GlebRadaev/shlink/internal/repository"
	"github.com/GlebRadaev/shlink/internal/service"
	"github.com/GlebRadaev/shlink/internal/taskmanager"
)

// Application is the main struct that encapsulates the application context, configurations, services, server, and worker pool.
type Application struct {
	Ctx        context.Context
	Config     *config.Config
	Logger     *logger.Logger
	Services   *service.Services
	Server     *http.Server
	WorkerPool *taskmanager.WorkerPool
}

// NewApplication creates a new instance of Application with the provided context.
func NewApplication(ctx context.Context) *Application {
	return &Application{Ctx: ctx}
}

// Init initializes the application by loading configurations, setting up services, and preparing the server and router.
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

	if app.Config.EnableHTTPS {
		err := GenerateCertificate(app.Config.CertPath, app.Config.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to generate certificates: %v", err)
		}
	}

	app.WorkerPool = taskmanager.NewWorkerPool(app.Ctx, 100, 10)
	repositories := repository.NewRepositoryFactory(app.Ctx, app.Config, app.Logger)
	app.Services = service.NewServiceFactory(app.Ctx, app.Config, app.Logger, app.WorkerPool, repositories)
	router := app.SetupRoutes()

	app.Server = &http.Server{
		Addr:    app.Config.ServerAddress,
		Handler: router,
	}
	return nil
}

// Start launches the HTTP server and listens for incoming requests.
func (app *Application) Start() error {
	go func() {
		logger := app.Logger.Named("Server Initialization")
		logger.Infoln("Server started at", app.Config.ServerAddress)
		logger.Infoln("Base URL:", app.Config.BaseURL)
		logger.Infoln("File storage path:", app.Config.FileStoragePath)
		logger.Infoln("Database path:", app.Config.DatabaseDSN)
		if app.Config.EnableHTTPS {
			logger.Infoln("Starting server with HTTPS...")
			if err := app.Server.ListenAndServeTLS(app.Config.CertPath, app.Config.KeyPath); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("HTTPS server error: %v", err)
			}
		} else {
			logger.Infoln("Starting server with HTTP...")
			if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("HTTP server error: %v", err)
			}
		}
	}()
	<-app.Ctx.Done()
	return app.Shutdown()
}

// Shutdown gracefully shuts down the server, saves data, and stops the worker pool.
func (app *Application) Shutdown() error {
	logger := app.Logger.Named("Server Shutdown")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Error during server shutdown: %v", err)
	} else {
		logger.Info("Server shutdown successfully")
	}
	saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer saveCancel()
	if err := app.Services.URLService.SaveData(saveCtx); err != nil {
		logger.Errorf("Failed to save data: %v", err)
	} else {
		logger.Info("Data successfully saved before shutdown")
	}
	app.WorkerPool.Shutdown()
	logger.Info("Worker pool shutdown completed")
	return nil
}

// SetupRoutes sets up the HTTP routes for the application.
func (app *Application) SetupRoutes() *chi.Mux {
	router := chi.NewRouter()
	middleware.Middleware(router)
	urlHandlers := handlers.NewURLHandlers(app.Services.URLService)
	healthHandlers := handlers.NewHealthHandlers(app.Services.HealthService)
	api.Routes(router, urlHandlers, healthHandlers)
	return router
}
