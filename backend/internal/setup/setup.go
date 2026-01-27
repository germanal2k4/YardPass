package setup

import (
	"errors"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"yardpass/internal/api"
	"yardpass/internal/api/handlers"
	"yardpass/internal/auth"
	"yardpass/internal/config"
	"yardpass/internal/domain"
	"yardpass/internal/observability/logger"
	"yardpass/internal/qr"
	"yardpass/internal/redis"
	"yardpass/internal/repo"
	"yardpass/internal/service"
	"yardpass/internal/telegram"
)

func SetupApi(configPath string) (*fx.App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if cfg.JWT.Secret == "" {
		return nil, errors.New("JWT_SECRET is required for API server")
	}

	return fx.New(
		fx.StartTimeout(cfg.Server.StartTimeout),
		fx.StopTimeout(cfg.Server.StopTimeout),
		fx.Provide(
			logger.NewLogger,

			repo.NewPostgresRepo,
			fx.Annotate(repo.NewPassRepo, fx.As(new(domain.PassRepository))),
			fx.Annotate(repo.NewApartmentRepo, fx.As(new(domain.ApartmentRepository))),
			fx.Annotate(repo.NewBuildingRepo, fx.As(new(domain.BuildingRepository))),
			fx.Annotate(repo.NewRuleRepo, fx.As(new(domain.RuleRepository))),
			fx.Annotate(repo.NewUserRepo, fx.As(new(domain.UserRepository))),
			fx.Annotate(repo.NewResidentRepo, fx.As(new(domain.ResidentRepository))),
			fx.Annotate(repo.NewScanEventRepo, fx.As(new(domain.ScanEventRepository))),

			redis.NewClient,

			auth.NewJWTService,
			service.NewPassService,
			service.NewUserService,
			service.NewResidentService,

			handlers.NewAuthHandler,
			handlers.NewPassHandler,
			handlers.NewRuleHandler,
			handlers.NewUserHandler,
			handlers.NewResidentHandler,
			handlers.NewScanEventHandler,
			handlers.NewReportHandler,
			handlers.NewParkingHandler,

			api.NewRouter,

			func() *config.Config { return cfg },
			func() config.PGConfig { return cfg.PG },
			func() config.RedisConfig { return cfg.Redis },
			func() config.JWTConfig { return cfg.JWT },
			func() config.LogConfig { return cfg.Log },
		),

		fx.Invoke(func(logger *zap.Logger) {}),
		fx.Invoke(func(repo *repo.PostgresRepo) {}),
		fx.Invoke(func(redisClient *redis.Client) {}),
		fx.Invoke(func(router *api.Router) {}),
	), nil
}

func SetupBot(configPath string) (*fx.App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if cfg.Telegram.BotToken == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN is required for Telegram bot")
	}

	return fx.New(
		fx.StartTimeout(cfg.Server.StartTimeout),
		fx.StopTimeout(cfg.Server.StopTimeout),
		fx.Provide(
			logger.NewLogger,

			repo.NewPostgresRepo,
			fx.Annotate(repo.NewPassRepo, fx.As(new(domain.PassRepository))),
			fx.Annotate(repo.NewApartmentRepo, fx.As(new(domain.ApartmentRepository))),
			fx.Annotate(repo.NewRuleRepo, fx.As(new(domain.RuleRepository))),
			fx.Annotate(repo.NewResidentRepo, fx.As(new(domain.ResidentRepository))),
			fx.Annotate(repo.NewScanEventRepo, fx.As(new(domain.ScanEventRepository))),

			redis.NewClient,

			service.NewPassService,
			qr.NewGenerator,

			telegram.NewBot,

			func() *config.Config { return cfg },
			func() config.PGConfig { return cfg.PG },
			func() config.RedisConfig { return cfg.Redis },
			func() config.LogConfig { return cfg.Log },
		),

		fx.Invoke(func(logger *zap.Logger) {}),
		fx.Invoke(func(repo *repo.PostgresRepo) {}),
		fx.Invoke(func(redisClient *redis.Client) {}),
		fx.Invoke(func(bot *telegram.Bot) {}),
	), nil
}
