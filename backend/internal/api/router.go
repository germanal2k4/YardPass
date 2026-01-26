package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"yardpass/internal/api/handlers"
	"yardpass/internal/api/middleware"
	"yardpass/internal/auth"
	"yardpass/internal/config"
	"yardpass/internal/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Router struct {
	cfg    *config.Config
	srv    *http.Server
	logger *zap.Logger
	router *gin.Engine
}

func NewRouter(
	lf fx.Lifecycle,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	passHandler *handlers.PassHandler,
	ruleHandler *handlers.RuleHandler,
	userHandler *handlers.UserHandler,
	residentHandler *handlers.ResidentHandler,
	scanEventHandler *handlers.ScanEventHandler,
	reportHandler *handlers.ReportHandler,
	parkingHandler *handlers.ParkingHandler,
	jwtService *auth.JWTService,
	redisClient *redis.Client,
	logger *zap.Logger,
) *Router {
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.InMemoryRateLimit(100, 200))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(jwtService))
	{
		api.GET("/me", authHandler.Me)

		passes := api.Group("/passes")
		{
			passes.POST("", middleware.CreatePassRateLimit(redisClient, cfg.RateLimit.CreatePassPerHour, time.Hour), passHandler.Create)
			passes.GET("/:id", passHandler.GetByID)
			passes.POST("/:id/revoke", passHandler.Revoke)
			passes.POST("/validate", passHandler.Validate)
			passes.GET("/active", passHandler.GetActive)
			passes.GET("/search", passHandler.Search)
		}

		rules := api.Group("/rules")
		rules.Use(middleware.RequireRole("admin", "superuser"))
		{
			rules.GET("", ruleHandler.Get)
			rules.PUT("", ruleHandler.Update)
		}

		users := api.Group("/users")
		users.Use(middleware.RequireRole("admin", "superuser"))
		{
			users.POST("", userHandler.RegisterUser)
			users.GET("", userHandler.ListUsers)
		}

		residents := api.Group("/residents")
		residents.Use(middleware.RequireRole("admin", "superuser"))
		{
			residents.POST("", residentHandler.CreateResident)
			residents.POST("/bulk", residentHandler.BulkCreateResidents)
			residents.POST("/import", residentHandler.ImportFromCSV)
			residents.GET("", residentHandler.ListResidents)
			residents.DELETE("/:id", residentHandler.DeleteResident)
		}

		scanEvents := api.Group("/scan-events")
		scanEvents.Use(middleware.RequireRole("guard", "admin", "superuser"))
		{
			scanEvents.GET("", scanEventHandler.ListEvents)
		}

		reports := api.Group("/reports")
		reports.Use(middleware.RequireRole("admin", "superuser"))
		{
			reports.GET("/statistics", reportHandler.GetStatistics)
			reports.GET("/export", reportHandler.ExportToExcel)
		}

		parking := api.Group("/parking")
		parking.Use(middleware.RequireRole("guard", "admin", "superuser"))
		{
			parking.GET("/occupancy", parkingHandler.GetOccupancy)
			parking.GET("/vehicles", parkingHandler.GetVehicles)
		}
	}

	service := r.Group("/service/v1")
	service.Use(middleware.ServiceAuthMiddleware(cfg.Service.Token))
	{
		service.POST("/passes", middleware.CreatePassRateLimit(redisClient, cfg.RateLimit.CreatePassPerHour, time.Hour), passHandler.Create)
		service.POST("/passes/:id/revoke", passHandler.Revoke)
		service.GET("/passes/active", passHandler.GetActive)
	}

	router := &Router{
		cfg:    cfg,
		logger: logger,
		router: r,
	}

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return router.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return router.Stop(ctx)
		},
	})

	return router
}

func (r *Router) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%s", r.cfg.Server.Host, r.cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		r.logger.Info("API server listening", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	r.srv = srv

	return nil
}

func (r *Router) Stop(ctx context.Context) error {
	if err := r.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	r.logger.Info("Server exited")
	return nil
}
