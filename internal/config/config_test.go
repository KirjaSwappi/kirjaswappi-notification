package config

import (
	"os"
	"reflect"
	"testing"
)

func TestGetEnvAsSlice(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue []string
		expected     []string
	}{
		{
			name:         "Environment variable set with multiple values",
			envKey:       "TEST_SLICE_MULTI",
			envValue:     "val1,val2,val3",
			defaultValue: []string{"default"},
			expected:     []string{"val1", "val2", "val3"},
		},
		{
			name:         "Environment variable set with single value",
			envKey:       "TEST_SLICE_SINGLE",
			envValue:     "val1",
			defaultValue: []string{"default"},
			expected:     []string{"val1"},
		},
		{
			name:         "Environment variable not set",
			envKey:       "TEST_SLICE_NOT_SET",
			envValue:     "",
			defaultValue: []string{"default"},
			expected:     []string{"default"},
		},
		{
			name:         "Environment variable set with empty values",
			envKey:       "TEST_SLICE_EMPTY_VALS",
			envValue:     "val1,,val2",
			defaultValue: []string{"default"},
			expected:     []string{"val1", "val2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.envKey, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.envKey); err != nil {
						t.Errorf("Failed to unset env var: %v", err)
					}
				}()
			}

			got := getEnvAsSlice(tt.envKey, tt.defaultValue)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("getEnvAsSlice() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLoad_StrictRequiresOrigins(t *testing.T) {
	t.Setenv("NOTIFY_STRICT", "true")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("API_KEY", "k")
	t.Setenv("JWT_SECRET", "s")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when ALLOWED_ORIGINS empty in strict mode")
	}
}

func TestLoad_StrictRejectsWildcardOrigin(t *testing.T) {
	t.Setenv("NOTIFY_STRICT", "yes")
	t.Setenv("ALLOWED_ORIGINS", "*")
	t.Setenv("API_KEY", "k")
	t.Setenv("JWT_SECRET", "s")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for wildcard ALLOWED_ORIGINS in strict mode")
	}
}

func TestLoad_StrictRequiresAPIKey(t *testing.T) {
	t.Setenv("NOTIFY_STRICT", "1")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")
	t.Setenv("API_KEY", "")
	t.Setenv("JWT_SECRET", "s")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when API_KEY empty in strict mode")
	}
}

func TestLoad_StrictRequiresJWTSecret(t *testing.T) {
	t.Setenv("NOTIFY_STRICT", "on")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")
	t.Setenv("API_KEY", "k")
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_SECRET empty in strict mode")
	}
}

func TestLoad_NonStrictDefaultOrigins(t *testing.T) {
	t.Setenv("NOTIFY_STRICT", "false")
	t.Setenv("ALLOWED_ORIGINS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"http://localhost:5173"}
	if !reflect.DeepEqual(cfg.AllowedOrigins, want) {
		t.Fatalf("AllowedOrigins=%v want %v", cfg.AllowedOrigins, want)
	}
}

func TestLoad_EnableGRPCReflectionParsing(t *testing.T) {
	t.Setenv("NOTIFY_STRICT", "false")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("ENABLE_GRPC_REFLECTION", "no")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.EnableGRPCReflection {
		t.Fatal("expected EnableGRPCReflection=false")
	}
	t.Setenv("ENABLE_GRPC_REFLECTION", "true")
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg2.EnableGRPCReflection {
		t.Fatal("expected EnableGRPCReflection=true")
	}
}
