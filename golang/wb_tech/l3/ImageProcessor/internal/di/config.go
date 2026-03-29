package di

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wb-go/wbf/retry"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
)

type Config struct {
	HTTP       HTTPConfig
	Kafka      KafkaConfig
	Storage    StorageConfig
	Processing ProcessingConfig
	Web        WebConfig
}

type HTTPConfig struct {
	Host string
	Port string
}

type KafkaConfig struct {
	Brokers           []string
	Topic             string
	Group             string
	Partitions        int
	ReplicationFactor int
	RetryAttempt      int
	RetryDelay        time.Duration
	RetryBackoff      int
}

type StorageConfig struct {
	RootDir string
}

type ProcessingConfig struct {
	MaxWidth        int
	MaxHeight       int
	ThumbnailSize   int
	WatermarkText   string
	MaxUploadSizeMB int
}

type WebConfig struct {
	Dir string
}

func ReadConfig() Config {
	return Config{
		HTTP: HTTPConfig{
			Host: getEnv("HTTP_HOST", "0.0.0.0"),
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Kafka: KafkaConfig{
			Brokers:           parseList(getEnv("KAFKA_BROKERS", "localhost:9092")),
			Topic:             getEnv("KAFKA_TOPIC", "image-processing"),
			Group:             getEnv("KAFKA_GROUP", "image-processor"),
			Partitions:        getEnvInt("KAFKA_TOPIC_PARTITIONS", 1),
			ReplicationFactor: getEnvInt("KAFKA_TOPIC_REPLICATION_FACTOR", 1),
			RetryAttempt:      getEnvInt("KAFKA_RETRY_ATTEMPTS", 3),
			RetryDelay:        time.Duration(getEnvInt("KAFKA_RETRY_DELAY_SECONDS", 2)) * time.Second,
			RetryBackoff:      getEnvInt("KAFKA_RETRY_BACKOFF", 2),
		},
		Storage: StorageConfig{
			RootDir: getEnv("STORAGE_ROOT_DIR", "storage"),
		},
		Processing: ProcessingConfig{
			MaxWidth:        getEnvInt("PROCESSING_MAX_WIDTH", 1280),
			MaxHeight:       getEnvInt("PROCESSING_MAX_HEIGHT", 1280),
			ThumbnailSize:   getEnvInt("PROCESSING_THUMBNAIL_SIZE", 320),
			WatermarkText:   getEnv("PROCESSING_WATERMARK_TEXT", "ImageProcessor"),
			MaxUploadSizeMB: getEnvInt("UPLOAD_MAX_SIZE_MB", 10),
		},
		Web: WebConfig{
			Dir: getEnv("WEB_DIR", "web"),
		},
	}
}

func (c KafkaConfig) RetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: c.RetryAttempt,
		Delay:    c.RetryDelay,
		Backoff:  float64(c.RetryBackoff),
	}
}

func (c ProcessingConfig) Options() entity.ProcessingOptions {
	return entity.ProcessingOptions{
		MaxWidth:      c.MaxWidth,
		MaxHeight:     c.MaxHeight,
		ThumbnailSize: c.ThumbnailSize,
		WatermarkText: c.WatermarkText,
	}
}

func (c ProcessingConfig) MaxUploadSizeBytes() int64 {
	return int64(c.MaxUploadSizeMB) * 1024 * 1024
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
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

func parseList(value string) []string {
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
