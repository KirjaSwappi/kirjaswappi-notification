package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort              int
	GRPCPort              int
	LogLevel              string
	AllowedOrigins        []string
	ShutdownTimeout       int // seconds
	APIKey                string
	JWTSecret             string
	EnableGRPCReflection  bool
}

// Load reads configuration from the environment. Returns an error if the
// process is asked to run in a strict (production) configuration but is
// missing required variables. Strict mode is opted in via NOTIFY_STRICT=true.
func Load() (*Config, error) {
	strict := getEnv("NOTIFY_STRICT", "") == "true"

	rawOrigins := os.Getenv("ALLOWED_ORIGINS")
	allowed := parseCSV(rawOrigins)

	if strict {
		if len(allowed) == 0 {
			return nil, fmt.Errorf("ALLOWED_ORIGINS must be set in strict mode")
		}
		for _, o := range allowed {
			if o == "*" {
				return nil, fmt.Errorf("ALLOWED_ORIGINS=* is not permitted in strict mode")
			}
		}
		if getEnv("API_KEY", "") == "" {
			return nil, fmt.Errorf("API_KEY must be set in strict mode")
		}
		if getEnv("JWT_SECRET", "") == "" {
			return nil, fmt.Errorf("JWT_SECRET must be set in strict mode")
		}
	}

	if len(allowed) == 0 {
		// Dev-friendly default that does NOT match `*` so misconfigured prod
		// deploys at least default to localhost only.
		allowed = []string{"http://localhost:5173"}
	}

	return &Config{
		HTTPPort:             getEnvAsInt("HTTP_PORT", 8080),
		GRPCPort:             getEnvAsInt("GRPC_PORT", 50051),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		AllowedOrigins:       allowed,
		ShutdownTimeout:      getEnvAsInt("SHUTDOWN_TIMEOUT", 30),
		APIKey:               getEnv("API_KEY", ""),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		EnableGRPCReflection: getEnvAsBool("ENABLE_GRPC_REFLECTION", false),
	}, nil
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
		result := parseCSV(value)
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func parseCSV(value string) []string {
	if value == "" {
		return nil
	}
	result := []string{}
	for _, v := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnvAsBool(key string, defaultValue bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}
