package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/observability"
	"yardpass/internal/qr"
	"yardpass/internal/redis"
	"yardpass/internal/repo"
	"yardpass/internal/service"
	"yardpass/internal/telegram"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Telegram.BotToken == "" {
		fmt.Fprintf(os.Stderr, "TELEGRAM_BOT_TOKEN is required for Telegram bot\n")
		os.Exit(1)
	}

	logger, err := observability.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting YardPass Telegram bot")

	ctx := context.Background()
	pool, err := repo.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	postgresRepo := repo.NewPostgresRepo(pool, logger)
	passRepo := repo.NewPassRepo(postgresRepo)
	apartmentRepo := repo.NewApartmentRepo(postgresRepo)
	residentRepo := repo.NewResidentRepo(postgresRepo)
	ruleRepo := repo.NewRuleRepo(postgresRepo)
	scanEventRepo := repo.NewScanEventRepo(postgresRepo)

	redisClient, err := redis.NewClient(cfg.Redis.URL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	passService := service.NewPassService(passRepo, apartmentRepo, ruleRepo, scanEventRepo, logger)
	qrGen := qr.NewGenerator()

	bot := telegram.NewBot(
		cfg,
		passService,
		residentRepo,
		apartmentRepo,
		qrGen,
		redisClient,
		logger,
	)

	if cfg.Telegram.WebhookURL != "" {
		logger.Info("Setting up webhook", zap.String("url", cfg.Telegram.WebhookURL))
		setupWebhook(bot, cfg, logger)
	} else {
		logger.Info("Starting polling mode")
		startPolling(bot, logger)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down bot...")
	logger.Info("Bot exited")
}

func setupWebhook(bot *telegram.Bot, cfg *config.Config, logger *zap.Logger) {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var update telegram.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			logger.Error("Failed to decode update", zap.Error(err))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		bot.ProcessUpdate(r.Context(), update)
		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf("%s:8081", cfg.Server.Host)
	logger.Info("Webhook server listening", zap.String("address", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal("Failed to start webhook server", zap.Error(err))
	}
}

func startPolling(bot *telegram.Bot, logger *zap.Logger) {
	ctx := context.Background()
	offset := int64(0)

	for {
		updates, err := bot.GetUpdates(ctx, offset)
		if err != nil {
			logger.Error("Failed to get updates", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			bot.ProcessUpdate(ctx, update)
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
		}

		time.Sleep(1 * time.Second)
	}
}
