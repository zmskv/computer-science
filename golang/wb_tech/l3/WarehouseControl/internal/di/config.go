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
	Auth     AuthConfig
	Web      WebConfig
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

type AuthConfig struct {
	JWTSecret     string
	TokenTTLHours int
}

type WebConfig struct {
	Dir string
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
			User:                   env.String("POSTGRES_USER", "warehouse"),
			Password:               env.String("POSTGRES_PASSWORD", "warehouse"),
			Name:                   env.String("POSTGRES_DB", "warehouse"),
			SSLMode:                env.String("POSTGRES_SSLMODE", "disable"),
			SlaveDSNs:              env.List("POSTGRES_SLAVE_DSNS"),
			MaxOpenConns:           env.Int("POSTGRES_MAX_OPEN_CONNS", 10),
			MaxIdleConns:           env.Int("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetimeSeconds: env.Int("POSTGRES_CONN_MAX_LIFETIME_SECONDS", 300),
		},
		Auth: AuthConfig{
			JWTSecret:     env.String("JWT_SECRET", "warehouse-control-secret"),
			TokenTTLHours: env.Int("JWT_TTL_HOURS", 12),
		},
		Web: WebConfig{
			Dir: env.String("WEB_DIR", "web"),
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

func (c AuthConfig) TokenTTL() time.Duration {
	return time.Duration(c.TokenTTLHours) * time.Hour
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
