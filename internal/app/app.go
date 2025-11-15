package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"delayed-notifier/internal/broker/rabbitmq"
	"delayed-notifier/internal/config"
	"delayed-notifier/internal/handler"
	"delayed-notifier/internal/repository/postgres"
	"delayed-notifier/internal/repository/redis"
	"delayed-notifier/internal/usecase"
	"delayed-notifier/internal/usecase/notifier"

	"github.com/wb-go/wbf/dbpg"
	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
)

const ShutdownTimeout = 5 * time.Second

type App struct {
	cfg      *config.Config
	db       *dbpg.DB
	rd       *wbfredis.Client
	broker   *rabbitmq.Broker
	repo     usecase.NotificationRepository
	notifier usecase.Notifier
	uc       handler.NotificationService
	server   *http.Server
	worker   *rabbitmq.Worker
}

func NewApp(cfg *config.Config) *App {
	retries := cfg.DefaultRetryStrategy()

	dbOpts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}
	db, err := dbpg.New(cfg.DBDSN(), cfg.DB.Slaves, dbOpts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	rd := wbfredis.New(cfg.RedisAddr(), cfg.Redis.Pass, cfg.Redis.DB)
	cache := redis.NewRedisCache(rd, retries)
	repo := postgres.NewNotificationRepository(db, cache, retries, time.Duration(cfg.CacheTTLHours)*time.Hour)
	br := rabbitmq.NewRabbitMQ(cfg, retries)
	notifier := notifier.NewMultiNotifier(cfg)
	uc := usecase.NewNotificationUsecase(repo, br, retries, notifier)
	worker := rabbitmq.NewWorker(br, uc.ProcessNotification)
	h := handler.NewHandler(uc)
	mux := handler.SetupRouter(h)
	muxWithMw := handler.LoggingMiddleware(mux)
	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: muxWithMw,
	}
	return &App{
		cfg:      cfg,
		db:       db,
		rd:       rd,
		broker:   br,
		repo:     repo,
		uc:       uc,
		notifier: notifier,
		server:   srv,
		worker:   worker,
	}
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.worker.Start(ctx)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Fatal().Err(err).Msg("Server failed")
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zlog.Logger.Info().Msg("Shutting down...")
	cancel()
	a.worker.Stop()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancelShutdown()
	if err := a.server.Shutdown(ctxShutdown); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Shutdown failed")
	}
	a.broker.Close()
	a.db.Master.Close()
	a.rd.Close()
	zlog.Logger.Info().Msg("Exited")
}
