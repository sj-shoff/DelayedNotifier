package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type Config struct {
	DB struct {
		Host            string        `env:"POSTGRES_HOST" validate:"required"`
		Port            int           `env:"POSTGRES_PORT" validate:"required"`
		User            string        `env:"POSTGRES_USER" validate:"required"`
		Pass            string        `env:"POSTGRES_PASSWORD" validate:"required"`
		DBName          string        `env:"POSTGRES_DB" validate:"required"`
		MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"`
		MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS"`
		ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME"`
		Slaves          []string      `env:"DB_SLAVES"`
	}
	Redis struct {
		Host string `env:"REDIS_HOST" validate:"required"`
		Port int    `env:"REDIS_PORT" validate:"required"`
		Pass string `env:"REDIS_PASSWORD"`
		DB   int    `env:"REDIS_DB"`
	}
	RabbitMQ struct {
		Host           string        `env:"RABBITMQ_HOST" validate:"required"`
		Port           int           `env:"RABBITMQ_PORT" validate:"required"`
		User           string        `env:"RABBITMQ_USER" validate:"required"`
		Pass           string        `env:"RABBITMQ_PASSWORD" validate:"required"`
		VHost          string        `env:"RABBITMQ_VHOST"`
		ConnectTimeout time.Duration `env:"RABBITMQ_CONNECT_TIMEOUT"`
		Heartbeat      time.Duration `env:"RABBITMQ_HEARTBEAT"`
	}
	Server struct {
		Addr         string        `env:"SERVER_PORT" validate:"required"`
		ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" validate:"required"`
		WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" validate:"required"`
		IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT" validate:"required"`
	}
	Retries struct {
		Attempts int     `env:"RETRIES_ATTEMPTS" validate:"required,gte=1"`
		DelayMs  int     `env:"RETRIES_DELAY_MS" validate:"required,gte=0"`
		Backoff  float64 `env:"RETRIES_BACKOFF" validate:"required,gte=1"`
	}
	CacheTTLHours int `env:"CACHE_TTL_HOURS" validate:"required,gte=1"`
	Email         Email
	Telegram      Telegram
}

type Email struct {
	SmtpHost string `env:"EMAIL_SMTP_HOST"`
	SmtpPort int    `env:"EMAIL_SMTP_PORT"`
	User     string `env:"EMAIL_USER"`
	Pass     string `env:"EMAIL_PASSWORD"`
}

type Telegram struct {
	BotToken string `env:"TELEGRAM_BOT_TOKEN"`
}

func MustLoad() (*Config, error) {
	var cfg Config

	// Сначала пробуем загрузить из .env файла
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		// Если .env файла нет, загружаем из переменных окружения
		zlog.Logger.Warn().Err(err).Msg("Failed to load .env file, using environment variables")
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to read environment variables: %w", err)
		}
	}

	// Дополнительная валидация с помощью validator
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	zlog.Logger.Info().Msg("Configuration loaded and validated successfully")
	return &cfg, nil
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.DBName)
}

func (c *Config) RabbitMQDSN() string {
	vhost := c.RabbitMQ.VHost
	if vhost == "" {
		vhost = "/"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.RabbitMQ.User, c.RabbitMQ.Pass, c.RabbitMQ.Host, c.RabbitMQ.Port, vhost)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) DefaultRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: c.Retries.Attempts,
		Delay:    time.Duration(c.Retries.DelayMs) * time.Millisecond,
		Backoff:  c.Retries.Backoff,
	}
}
