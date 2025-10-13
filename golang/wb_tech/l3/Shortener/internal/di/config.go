package di

import (
	"fmt"
	"os"
	"strings"
)

type HTTPConfig struct {
	Host string
	Port string
}

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
}

func ReadConfig() Config {
	return Config{
		HTTP: HTTPConfig{
			Host: getEnv("HTTP_HOST", "0.0.0.0"),
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Postgres: PostgresConfig{
			MasterDSN: getEnv("POSTGRES_MASTER_DSN", ""),
			SlaveDSNs: parseSlaveDSNs(getEnv("POSTGRES_SLAVE_DSNS", "")),
			Host:      getEnv("POSTGRES_HOST", "localhost"),
			Port:      getEnv("POSTGRES_PORT", "5432"),
			Database:  getEnv("POSTGRES_DB", "shortener_db"),
			User:      getEnv("POSTGRES_USER", "shortener_user"),
			Password:  getEnv("POSTGRES_PASSWORD", "shortener_pass"),
		},
	}
}

type PostgresConfig struct {
	MasterDSN string
	SlaveDSNs []string
	Host      string
	Port      string
	Database  string
	User      string
	Password  string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseSlaveDSNs(slaveDSNs string) []string {
	if slaveDSNs == "" {
		return nil
	}
	return strings.Split(slaveDSNs, ",")
}

func (p *PostgresConfig) BuildMasterDSN() string {
	if p.MasterDSN != "" {
		return p.MasterDSN
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Database)
}
