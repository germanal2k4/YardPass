package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/domain"
	"yardpass/internal/qr"
	"yardpass/internal/redis"
	"yardpass/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Bot struct {
	token      string
	apiURL     string
	webhookURL string
	serverHost string
	serverPort string

	passService   *service.PassService
	residentRepo  domain.ResidentRepository
	apartmentRepo domain.ApartmentRepository
	qrGen         *qr.Generator
	redis         *redis.Client
	logger        *zap.Logger
	states        map[int64]*UserState
	location      *time.Location

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

type UserState struct {
	Step      string
	Data      map[string]interface{}
	ExpiresAt time.Time
}

const (
	StateWaitingGuestType  = "waiting_guest_type"
	StateWaitingCarPlate   = "waiting_car_plate"
	StateWaitingDuration   = "waiting_duration"
	StateWaitingCustomTime = "waiting_custom_time"
	StateWaitingGuestName  = "waiting_guest_name"
)

func NewBot(
	lf fx.Lifecycle,
	cfg *config.Config,
	passService *service.PassService,
	residentRepo domain.ResidentRepository,
	apartmentRepo domain.ApartmentRepository,
	qrGen *qr.Generator,
	redisClient *redis.Client,
	logger *zap.Logger,
) *Bot {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		logger.Warn("Failed to load Europe/Moscow timezone, using UTC", zap.Error(err))
		location = time.UTC
	}

	bot := &Bot{
		token:         cfg.Telegram.BotToken,
		apiURL:        fmt.Sprintf("https://api.telegram.org/bot%s", cfg.Telegram.BotToken),
		webhookURL:    cfg.Telegram.WebhookURL,
		serverHost:    cfg.Telegram.ServerHost,
		serverPort:    cfg.Telegram.ServerPort,
		passService:   passService,
		residentRepo:  residentRepo,
		apartmentRepo: apartmentRepo,
		qrGen:         qrGen,
		redis:         redisClient,
		logger:        logger,
		states:        make(map[int64]*UserState),
		location:      location,
	}

	lf.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return bot.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return bot.Stop(ctx)
		},
	})

	return bot
}

func (b *Bot) Start(ctx context.Context) error {
	if err := b.SetMyCommands(ctx); err != nil {
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	b.logger.Info("Bot commands menu set successfully")

	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.wg.Add(1)
	if b.webhookURL != "" {
		b.logger.Info("Setting up webhook", zap.String("url", b.webhookURL))
		go b.listenWebhook()
	} else {
		b.logger.Info("Starting polling mode")
		go b.startPolling(b.ctx)
	}

	return nil
}

func (b *Bot) Stop(ctx context.Context) error {
	b.logger.Info("Shutting down bot...")
	b.cancel()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		b.logger.Info("Bot exited")
		return nil
	case <-ctx.Done():
		b.logger.Info("Bot shutdown by context")
		return ctx.Err()
	}
}

func (b *Bot) listenWebhook() {
	defer b.wg.Done()
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var update Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			b.logger.Error("Failed to decode update", zap.Error(err))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		b.ProcessUpdate(r.Context(), update)
		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf("%s:%s", b.serverHost, b.serverPort)
	b.logger.Info("Webhook server listening", zap.String("address", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		b.logger.Fatal("Failed to start webhook server", zap.Error(err))
	}
}

func (b *Bot) startPolling(ctx context.Context) {
	defer b.wg.Done()
	offset := int64(0)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Polling mode stopped by context")
			return
		default:
		}

		updates, err := b.GetUpdates(ctx, offset)
		if err != nil {
			b.logger.Error("Failed to get updates", zap.Error(err))
			select {
			case <-ctx.Done():
				b.logger.Info("Polling mode stopped by context")
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		for _, update := range updates {
			b.ProcessUpdate(ctx, update)
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
		}

		select {
		case <-ctx.Done():
			b.logger.Info("Polling mode stopped by context")
			return
		case <-time.After(1 * time.Second):
		}
	}
}
