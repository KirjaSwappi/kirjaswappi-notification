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
