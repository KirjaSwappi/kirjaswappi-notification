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
}

func Load() *Config {
	return &Config{
		HTTPPort:        getEnvAsInt("HTTP_PORT", 8080),
		GRPCPort:        getEnvAsInt("GRPC_PORT", 50051),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		AllowedOrigins:  getEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}),
		ShutdownTimeout: getEnvAsInt("SHUTDOWN_TIMEOUT", 30),
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
		// Simple comma-separated parsing
		result := []string{}
		for _, v := range strings.Split(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
