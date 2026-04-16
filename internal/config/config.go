package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort        int
	GRPCPort        int
	LogLevel        string
	AllowedOrigins  []string
	ShutdownTimeout int // seconds
	APIKey          string
	JWTSecret       string
}

func Load() *Config {
	return &Config{
		HTTPPort:        getEnvAsInt("HTTP_PORT", 8080),
		GRPCPort:        getEnvAsInt("GRPC_PORT", 50051),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		AllowedOrigins:  getEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}),
		ShutdownTimeout: getEnvAsInt("SHUTDOWN_TIMEOUT", 30),
		APIKey:          getEnv("API_KEY", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		result := []string{}
		for _, v := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(v)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
