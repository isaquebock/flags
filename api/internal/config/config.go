package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	RedisAddr          string
	RedisPassword      string
	InternalToken      string
	AllowedOrigins     []string
	SnapshotMaxRetries int
	LogLevel           string
}

func LoadFromEnv() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		InternalToken: getEnv("INTERNAL_TOKEN", "change-me"),
		AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:5173")),
		SnapshotMaxRetries: getEnvInt("SNAPSHOT_MAX_RETRIES", 3),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if n, err := strconv.Atoi(value); err == nil {
		return n
	}
	return defaultValue
}

func parseOrigins(originsCsv string) []string {
	var origins []string
	for _, origin := range strings.Split(originsCsv, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
