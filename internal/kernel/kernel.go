package kernel

import (
	"context"
	"fmt"
	"go-inventory-reservations/internal/application"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/cron"
	"go-inventory-reservations/internal/database"
	"go-inventory-reservations/internal/handler"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/repository"
	"go-inventory-reservations/internal/router"
	"go-inventory-reservations/internal/service"
	"go-inventory-reservations/internal/uow"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	cronlib "github.com/robfig/cron/v3"
)

// Kernel is the main application kernel. It contains all required dependencies and handles the application lifecycle.
type Kernel struct {
	Config       *config.Config
	Logger       *logger.Logger
	DBConnection *database.Database
	Router       *gin.Engine
	HTTPServer   *http.Server
	Cron         *cronlib.Cron
}

// New creates and initializes a new Kernel instance.
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

	// Repositories
	unitOfWork := uow.NewUnitOfWorkManager(db.DB)
	stockRepo := repository.NewStockRepository(db.DB)
	reservationRepo := repository.NewReservationRepository(db.DB)
	reservationItemRepo := repository.NewReservationItemsRepository(db.DB)

	// Services
	stockService := service.NewStockService(stockRepo)
	reservationService := service.NewReservationService(reservationRepo, cfg)
	reservationItemService := service.NewReservationItemsService(reservationItemRepo, cfg)

	reservationOrchestrator := application.NewReservationOrchestrator(
		unitOfWork,
		stockService,
		reservationService,
		reservationItemService,
	)

	adminStockOrchestrator := application.NewAdminStockOrchestrator(
		unitOfWork,
		stockService,
		reservationItemService,
	)

	// Handlers (with all required services)
	handlersPool := handler.NewHandlersPool(reservationOrchestrator, adminStockOrchestrator, stockService, reservationService, log)

	// Gin router and HTTP server setup
	routerEngine := gin.Default()
	router.SetupRoutes(routerEngine, *handlersPool)

	port := strconv.Itoa(cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: routerEngine,
	}

	c := cron.InitCrons(reservationOrchestrator, reservationService, log, cfg)

	kernel := &Kernel{
		Config:       cfg,
		Logger:       log,
		DBConnection: db,
		Router:       routerEngine,
		HTTPServer:   httpServer,
		Cron:         c,
	}

	log.Info("Kernel initialized successfully")
	return kernel, nil
}

// Start starts the application kernel and the HTTP server (in a goroutine).
func (k *Kernel) Start(ctx context.Context) error {
	k.Logger.Info("Starting kernel...")

	if err := k.DBConnection.HealthCheck(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	k.Logger.Info("Database health check successful")

	go func() {
		k.Logger.Infof("🚀 Starting HTTP server on %s...", k.HTTPServer.Addr)
		if err := k.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			k.Logger.WithError(err).Error("Server failed")
		}
	}()

	k.Cron.Start()
	k.Logger.Info("Kernel started successfully")
	return nil
}

// Stop gracefully shuts down the HTTP server and closes the database connection.
func (k *Kernel) Stop(ctx context.Context) error {
	k.Logger.Info("Stopping kernel...")
	if k.Cron != nil {
		k.Cron.Stop()
	}

	// Graceful server shutdown with timeout
	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := k.HTTPServer.Shutdown(ctxShutdown); err != nil {
		k.Logger.WithError(err).Error("Graceful server shutdown failed")
	} else {
		k.Logger.Info("HTTP server stopped gracefully")
	}

	if err := k.DBConnection.Close(); err != nil {
		k.Logger.WithError(err).Error("Failed to close database connection")
	} else {
		k.Logger.Info("Database connection closed")
	}

	k.Logger.Info("Kernel stopped successfully")
	return nil
}

// WaitForShutdown waits for a SIGINT or SIGTERM signal and then returns.
func (k *Kernel) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	k.Logger.WithField("signal", sig.String()).Info("Received shutdown signal")
}

// Run starts the kernel, waits for a shutdown signal, then stops gracefully.
func (k *Kernel) Run() error {
	ctx := context.Background()

	if err := k.Start(ctx); err != nil {
		return err
	}

	k.WaitForShutdown()

	return k.Stop(ctx)
}
