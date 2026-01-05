package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yardpass/internal/api"
	"yardpass/internal/api/handlers"
	"yardpass/internal/auth"
	"yardpass/internal/config"
	"yardpass/internal/observability"
	"yardpass/internal/redis"
	"yardpass/internal/repo"
	"yardpass/internal/service"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Проверяем JWT_SECRET для API сервера
	if cfg.JWT.Secret == "" {
		fmt.Fprintf(os.Stderr, "JWT_SECRET is required for API server\n")
		os.Exit(1)
	}

	logger, err := observability.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting YardPass API server")

	ctx := context.Background()
	pool, err := repo.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	postgresRepo := repo.NewPostgresRepo(pool, logger)
	passRepo := repo.NewPassRepo(postgresRepo)
	apartmentRepo := repo.NewApartmentRepo(postgresRepo)
	buildingRepo := repo.NewBuildingRepo(postgresRepo)
	ruleRepo := repo.NewRuleRepo(postgresRepo)
	userRepo := repo.NewUserRepo(postgresRepo)
	residentRepo := repo.NewResidentRepo(postgresRepo)
	scanEventRepo := repo.NewScanEventRepo(postgresRepo)

	redisClient, err := redis.NewClient(cfg.Redis.URL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, userRepo)
	passService := service.NewPassService(passRepo, apartmentRepo, ruleRepo, scanEventRepo, logger)
	userService := service.NewUserService(userRepo, buildingRepo, logger)
	residentService := service.NewResidentService(residentRepo, apartmentRepo, logger)

	authHandler := handlers.NewAuthHandler(jwtService)
	passHandler := handlers.NewPassHandler(passService)
	ruleHandler := handlers.NewRuleHandler(ruleRepo)
	userHandler := handlers.NewUserHandler(userService)
	residentHandler := handlers.NewResidentHandler(residentService)
	scanEventHandler := handlers.NewScanEventHandler(scanEventRepo)
	reportHandler := handlers.NewReportHandler(scanEventRepo, passRepo)
	parkingHandler := handlers.NewParkingHandler(passService)

	router := api.SetupRouter(cfg, authHandler, passHandler, ruleHandler, userHandler, residentHandler, scanEventHandler, reportHandler, parkingHandler, jwtService, redisClient)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("API server listening", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
