package di

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Redis      RedisConfig
	HTTP       HTTPConfig
	RabbitMQ   RabbitMQConfig
	SMTPClient SMTPClient
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type HTTPConfig struct {
	Host string
	Port string
}

type RabbitMQConfig struct {
	URL      string
	Queue    string
	Exchange string
	Retries  int
	Pause    time.Duration
}

type SMTPClient struct {
	Email    string
	Password string
	Host     string
	Port     int
}

func ReadConfig() Config {
	redis := RedisConfig{
		Addr:     mustGetEnv("REDIS_HOST"),
		Password: mustGetEnv("REDIS_PASSWORD"),
		DB:       mustGetEnvInt("REDIS_DB"),
	}

	http := HTTPConfig{
		Host: mustGetEnv("HTTP_HOST"),
		Port: mustGetEnv("HTTP_PORT"),
	}

	rabbitmq := RabbitMQConfig{
		URL:      mustGetEnv("RABBITMQ_URL"),
		Queue:    mustGetEnv("RABBITMQ_QUEUE"),
		Exchange: mustGetEnv("RABBITMQ_EXCHANGE"),
		Retries:  mustGetEnvInt("RABBITMQ_RETRIES"),
		Pause:    time.Duration(mustGetEnvInt("RABBITMQ_PAUSE")) * time.Second,
	}

	smtp := SMTPClient{
		Host:     mustGetEnv("SMTP_HOST"),
		Port:     mustGetEnvInt("SMTP_PORT"),
		Email:    mustGetEnv("SMTP_EMAIL"),
		Password: mustGetEnv("SMTP_PASSWORD"),
	}

	return Config{
		Redis:      redis,
		HTTP:       http,
		RabbitMQ:   rabbitmq,
		SMTPClient: smtp,
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("environment variable %s is not set", key)
	}
	return val
}

func mustGetEnvInt(key string) int {
	valStr := mustGetEnv(key)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("invalid int value for %s: %v", key, err)
	}
	return val
}
