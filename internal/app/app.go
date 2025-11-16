// internal/app/app.go
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"delayed-notifier/internal/broker"
	"delayed-notifier/internal/broker/rabbitmq"
	"delayed-notifier/internal/config"
	"delayed-notifier/internal/domain"
	"delayed-notifier/internal/handler"
	"delayed-notifier/internal/repository"
	"delayed-notifier/internal/repository/postgres"
	"delayed-notifier/internal/repository/redis"
	"delayed-notifier/internal/usecase"
	"delayed-notifier/internal/usecase/notifier"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

const (
	ShutdownTimeout = 5 * time.Second
)

type App struct {
	cfg      *config.Config
	db       *dbpg.DB
	cache    repository.Cache
	broker   broker.Broker
	repo     repository.NotificationRepository
	notifier usecase.Notifier
	uc       handler.NotificationService
	server   *http.Server
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewApp(cfg *config.Config) (*App, error) {
	retries := cfg.DefaultRetryStrategy()

	// Инициализация базы данных
	dbOpts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}
	db, err := dbpg.New(cfg.DBDSN(), cfg.DB.Slaves, dbOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Инициализация кэша и репозиториев
	cache := redis.NewRedisCache(cfg, retries)
	repo := postgres.NewNotificationRepository(db, cache, retries, time.Duration(cfg.CacheTTLHours)*time.Hour)

	// Инициализация RabbitMQ брокера
	broker, err := rabbitmq.NewRabbitMQ(cfg, retries)
	if err != nil {
		db.Master.Close()
		cache.Close()
		return nil, fmt.Errorf("failed to create RabbitMQ broker: %w", err)
	}

	// Инициализация сервисов
	notifier := notifier.NewMultiNotifier(cfg)
	uc := usecase.NewNotificationUsecase(repo, broker, retries, notifier)

	// Инициализация HTTP handler
	h := handler.NewHandler(uc)
	mux := handler.SetupRouter(h)
	muxWithMw := handler.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: muxWithMw,
	}

	app := &App{
		cfg:      cfg,
		db:       db,
		cache:    cache,
		broker:   broker,
		repo:     repo,
		notifier: notifier,
		uc:       uc,
		server:   server,
	}

	// Запускаем consumer в Run, так как handler готов
	return app, nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	zlog.Logger.Info().Msg("Starting application...")

	// Создаем handler здесь
	handler := func(ctx context.Context, msg amqp091.Delivery) error {
		var payload struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to unmarshal message")
			return err
		}
		if payload.ID == "" {
			zlog.Logger.Error().Msg("Missing ID in payload")
			return domain.ErrNotFound
		}
		err := a.uc.ProcessNotification(ctx, payload.ID)
		if err != nil {
			zlog.Logger.Error().Err(err).Str("id", payload.ID).Msg("Failed to process notification")
			return err
		}
		zlog.Logger.Info().Str("id", payload.ID).Msg("Notification processed successfully")
		return nil
	}

	// Запускаем consumer
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.broker.Consume(ctx, "notifications", handler); err != nil && !errors.Is(err, context.Canceled) {
			zlog.Logger.Error().Err(err).Msg("Consumer stopped with error")
		}
	}()

	// Запускаем HTTP сервер
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		zlog.Logger.Info().Str("addr", a.server.Addr).Msg("Starting HTTP server")
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Logger.Error().Err(err).Msg("HTTP server failed")
			cancel()
		}
	}()

	// Ожидаем сигналы завершения
	a.waitForShutdown()
	return nil
}

// остальной код Shutdown и waitForShutdown как раньше
func (a *App) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	zlog.Logger.Info().Msg("Received shutdown signal")

	a.Shutdown()
}

func (a *App) Shutdown() {
	zlog.Logger.Info().Msg("Initiating graceful shutdown...")

	if a.cancel != nil {
		a.cancel()
	}

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancelShutdown()

	if err := a.server.Shutdown(ctxShutdown); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to shutdown HTTP server gracefully")
	}

	if a.broker != nil {
		if err := a.broker.Close(); err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to close broker connection")
		}
	}

	if a.db != nil {
		a.db.Master.Close()
	}

	if a.cache != nil {
		a.cache.Close()
	}

	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		zlog.Logger.Info().Msg("All components stopped gracefully")
	case <-time.After(ShutdownTimeout):
		zlog.Logger.Warn().Msg("Shutdown timeout exceeded, forcing exit")
	}

	zlog.Logger.Info().Msg("Application stopped")
}
