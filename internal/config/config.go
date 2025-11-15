package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type Config struct {
	DB struct {
		Host            string        `yaml:"host" env:"DB_HOST" validate:"required"`
		Port            int           `yaml:"port" env:"DB_PORT" validate:"required"`
		User            string        `yaml:"user" env:"DB_USER" validate:"required"`
		Pass            string        `yaml:"pass" env:"DB_PASS" validate:"required"`
		DBName          string        `yaml:"dbname" env:"DB_DBNAME" validate:"required"`
		MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" validate:"gte=0"`
		MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" validate:"gte=0"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" validate:"gte=0"`
		Slaves          []string      `yaml:"slaves" env:"DB_SLAVES"`
	} `yaml:"db"`
	Redis struct {
		Host string `yaml:"host" env:"REDIS_HOST" validate:"required"`
		Port int    `yaml:"port" env:"REDIS_PORT" validate:"required"`
		Pass string `yaml:"pass" env:"REDIS_PASS"`
		DB   int    `yaml:"db" env:"REDIS_DB"`
	} `yaml:"redis"`
	RabbitMQ struct {
		Host           string        `yaml:"host" env:"RABBITMQ_HOST" validate:"required"`
		Port           int           `yaml:"port" env:"RABBITMQ_PORT" validate:"required"`
		User           string        `yaml:"user" env:"RABBITMQ_USER" validate:"required"`
		Pass           string        `yaml:"pass" env:"RABBITMQ_PASS" validate:"required"`
		VHost          string        `yaml:"vhost" env:"RABBITMQ_VHOST"`
		ConnectTimeout time.Duration `yaml:"connect_timeout" env:"RABBITMQ_CONNECT_TIMEOUT" validate:"required"`
		Heartbeat      time.Duration `yaml:"heartbeat" env:"RABBITMQ_HEARTBEAT" validate:"required"`
	} `yaml:"rabbitmq"`
	Server struct {
		Addr         string        `yaml:"addr" env:"SERVER_ADDR" validate:"required"`
		ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" env-default:"10s" validate:"required,min=1000000000"`
		WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" env-default:"10s" validate:"required,min=1000000000"`
		IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"30s" validate:"required,min=1000000000"`
	} `yaml:"server"`
	Retries struct {
		Attempts int     `yaml:"attempts" env:"RETRIES_ATTEMPTS" validate:"gte=1"`
		DelayMs  int     `yaml:"delay_ms" env:"RETRIES_DELAY_MS" validate:"gte=0"` // Refactored: int для ms
		Backoff  float64 `yaml:"backoff" env:"RETRIES_BACKOFF" validate:"gte=1"`
	} `yaml:"retries"`
	CacheTTLHours int      `yaml:"cache_ttl_hours" env:"CACHE_TTL_HOURS" validate:"gte=1"`
	Email         Email    `yaml:"email"`
	Telegram      Telegram `yaml:"telegram"`
}

type Email struct {
	SmtpHost string `yaml:"smtp_host" env:"EMAIL_SMTP_HOST"`
	SmtpPort int    `yaml:"smtp_port" env:"EMAIL_SMTP_PORT"`
	User     string `yaml:"user" env:"EMAIL_USER"`
	Pass     string `yaml:"pass" env:"EMAIL_PASS"`
}

type Telegram struct {
	BotToken string `yaml:"bot_token" env:"TELEGRAM_BOT_TOKEN"`
}

func MustLoad() (*Config, error) {
	var c Config
	cfg := config.New()
	if err := cfg.LoadConfigFiles("config.yaml"); err != nil {
		zlog.Logger.Warn().Err(err).Msg("Failed to load config.yaml, using env")
	}
	if err := cfg.LoadEnvFiles(".env"); err != nil {
		zlog.Logger.Warn().Err(err).Msg("Failed to load .env")
	}
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.DBName)
}

func (c *Config) RabbitMQDSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.RabbitMQ.User, c.RabbitMQ.Pass, c.RabbitMQ.Host, c.RabbitMQ.Port, c.RabbitMQ.VHost)
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
