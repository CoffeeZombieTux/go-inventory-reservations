package kernel

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
	"go-inventory-reservations/internal/router"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/database"
	"go-inventory-reservations/internal/logger"
)

type Kernel struct {
	Config   *config.Config
	Logger   *logger.Logger
	Database *database.Database
	Server   *gin.Engine
}

func New() (*Kernel, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.New(cfg.Logger.Level, cfg.Logger.Format)
	log.Info("Starting application kernel")

	db, err := database.New(cfg.GetDatabaseDSN(), log.Logger)
	if err != nil {
		log.WithError(err).Error("Failed to initialize database")
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	handlers := handler.NewHandlers(log)
	engine := gin.Default()
	router.SetupRoutes(engine, *handlers)

	kernel := &Kernel{
		Config:   cfg,
		Logger:   log,
		Database: db,
		Server:   engine,
	}

	log.Info("Kernel initialized successfully")
	return kernel, nil
}

func (k *Kernel) Start(ctx context.Context) error {
	k.Logger.Info("Starting kernel...")

	if err := k.Database.HealthCheck(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	port := strconv.Itoa(k.Config.Server.Port)
	if err := k.Server.Run(":" + port); err != nil {
		return fmt.Errorf("server failed: %v", err)
	}
	k.Logger.Info("🚀 Starting server on port %s...", port)
	k.Logger.Info("Kernel started successfully")
	return nil
}

func (k *Kernel) Stop(ctx context.Context) error {
	k.Logger.Info("Stopping kernel...")

	if err := k.Database.Close(); err != nil {
		k.Logger.WithError(err).Error("Failed to close database connection")
	}

	k.Logger.Info("Kernel stopped successfully")
	return nil
}

func (k *Kernel) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	k.Logger.WithField("signal", sig.String()).Info("Received shutdown signal")
}

func (k *Kernel) Run() error {
	ctx := context.Background()

	if err := k.Start(ctx); err != nil {
		return err
	}

	k.WaitForShutdown()

	return k.Stop(ctx)
}
