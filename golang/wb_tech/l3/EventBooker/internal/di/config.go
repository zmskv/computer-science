package di

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wb-go/wbf/dbpg"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Worker   WorkerConfig
	Web      WebConfig
	Booking  BookingConfig
	Telegram TelegramConfig
}

type HTTPConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	URL                    string
	Host                   string
	Port                   string
	User                   string
	Password               string
	Name                   string
	SSLMode                string
	SlaveDSNs              []string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
}

type WorkerConfig struct {
	IntervalSeconds int
}

type WebConfig struct {
	Dir string
}

type BookingConfig struct {
	DefaultTTLMinutes int
}

type TelegramConfig struct {
	Enabled  bool
	BotToken string
	BaseURL  string
}

func ReadConfig() Config {
	env := envReader{}

	return Config{
		HTTP: HTTPConfig{
			Host: env.String("HTTP_HOST", "0.0.0.0"),
			Port: env.String("HTTP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL:                    env.String("DATABASE_URL", ""),
			Host:                   env.String("POSTGRES_HOST", "localhost"),
			Port:                   env.String("POSTGRES_PORT", "5432"),
			User:                   env.String("POSTGRES_USER", "eventbooker"),
			Password:               env.String("POSTGRES_PASSWORD", "eventbooker"),
			Name:                   env.String("POSTGRES_DB", "eventbooker"),
			SSLMode:                env.String("POSTGRES_SSLMODE", "disable"),
			SlaveDSNs:              env.List("POSTGRES_SLAVE_DSNS"),
			MaxOpenConns:           env.Int("POSTGRES_MAX_OPEN_CONNS", 10),
			MaxIdleConns:           env.Int("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetimeSeconds: env.Int("POSTGRES_CONN_MAX_LIFETIME_SECONDS", 300),
		},
		Worker: WorkerConfig{
			IntervalSeconds: env.Int("EXPIRATION_CHECK_INTERVAL_SECONDS", 5),
		},
		Web: WebConfig{
			Dir: env.String("WEB_DIR", "web"),
		},
		Booking: BookingConfig{
			DefaultTTLMinutes: env.Int("DEFAULT_BOOKING_TTL_MINUTES", 15),
		},
		Telegram: TelegramConfig{
			Enabled:  env.Bool("TELEGRAM_ENABLED", false),
			BotToken: env.String("TELEGRAM_BOT_TOKEN", ""),
			BaseURL:  env.String("TELEGRAM_BASE_URL", "https://api.telegram.org"),
		},
	}
}

func (c DatabaseConfig) MasterDSN() string {
	if c.URL != "" {
		return c.URL
	}

	encodedUser := url.QueryEscape(c.User)
	encodedPassword := url.QueryEscape(c.Password)

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		encodedUser,
		encodedPassword,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

func (c DatabaseConfig) Options() *dbpg.Options {
	return &dbpg.Options{
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: time.Duration(c.ConnMaxLifetimeSeconds) * time.Second,
	}
}

func (c WorkerConfig) Interval() time.Duration {
	seconds := c.IntervalSeconds
	if seconds <= 0 {
		seconds = 5
	}

	return time.Duration(seconds) * time.Second
}

func (c BookingConfig) DefaultTTL() time.Duration {
	minutes := c.DefaultTTLMinutes
	if minutes <= 0 {
		minutes = 15
	}

	return time.Duration(minutes) * time.Minute
}

type envReader struct{}

func (envReader) String(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func (envReader) Int(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func (envReader) Bool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	switch strings.ToLower(value) {
	case "true":
		return true
	case "false":
		return false
	default:
		return fallback
	}
}

func (env envReader) List(key string) []string {
	value := env.String(key, "")
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
