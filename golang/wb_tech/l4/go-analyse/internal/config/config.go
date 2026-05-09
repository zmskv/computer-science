package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultPort              = "8080"
	defaultGCPercent         = 100
	defaultShutdownTimeout   = 10 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultGinMode           = "release"
)

type Config struct {
	Port              string
	GCPercent         int
	GinMode           string
	ShutdownTimeout   time.Duration
	ReadHeaderTimeout time.Duration
}

func LoadFromEnv() Config {
	cfg := Config{
		Port:              defaultPort,
		GCPercent:         defaultGCPercent,
		GinMode:           defaultGinMode,
		ShutdownTimeout:   defaultShutdownTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	if value := os.Getenv("PORT"); value != "" {
		cfg.Port = value
	}

	if value := os.Getenv("GC_PERCENT"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			cfg.GCPercent = parsed
		}
	}

	if value := os.Getenv("GIN_MODE"); value != "" {
		switch value {
		case "debug", "release", "test":
			cfg.GinMode = value
		}
	}

	if value := os.Getenv("SHUTDOWN_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.ShutdownTimeout = parsed
		}
	}

	if value := os.Getenv("READ_HEADER_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.ReadHeaderTimeout = parsed
		}
	}

	return cfg
}

func (c Config) Address() string {
	return ":" + c.Port
}
