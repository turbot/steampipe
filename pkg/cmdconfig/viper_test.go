package cmdconfig

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestViper(t *testing.T) {
	v := Viper()
	if v == nil {
		t.Fatal("Viper() returned nil")
	}
	// Should return the global viper instance
	if v != viper.GetViper() {
		t.Error("Viper() should return the global viper instance")
	}
}

func TestSetBaseDefaults(t *testing.T) {
	// Save original viper state
	origTelemetry := viper.Get(pconstants.ArgTelemetry)
	origUpdateCheck := viper.Get(pconstants.ArgUpdateCheck)
	origPort := viper.Get(pconstants.ArgDatabasePort)
	defer func() {
		// Restore original state
		if origTelemetry != nil {
			viper.Set(pconstants.ArgTelemetry, origTelemetry)
		}
		if origUpdateCheck != nil {
			viper.Set(pconstants.ArgUpdateCheck, origUpdateCheck)
		}
		if origPort != nil {
			viper.Set(pconstants.ArgDatabasePort, origPort)
		}
	}()

	err := setBaseDefaults()
	if err != nil {
		t.Fatalf("setBaseDefaults() returned error: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{
			name:     "telemetry_default",
			key:      pconstants.ArgTelemetry,
			expected: constants.TelemetryInfo,
		},
		{
			name:     "update_check_default",
			key:      pconstants.ArgUpdateCheck,
			expected: true,
		},
		{
			name:     "database_port_default",
			key:      pconstants.ArgDatabasePort,
			expected: constants.DatabaseDefaultPort,
		},
		{
			name:     "autocomplete_default",
			key:      pconstants.ArgAutoComplete,
			expected: true,
		},
		{
			name:     "cache_enabled_default",
			key:      pconstants.ArgServiceCacheEnabled,
			expected: true,
		},
		{
			name:     "cache_max_ttl_default",
			key:      pconstants.ArgCacheMaxTtl,
			expected: 300,
		},
		{
			name:     "memory_max_mb_plugin_default",
			key:      pconstants.ArgMemoryMaxMbPlugin,
			expected: 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := viper.Get(tt.key)
			if val != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.key, val)
			}
		})
	}
}

func TestSetDefaultFromEnv_String(t *testing.T) {
	// Clean up viper state
	viper.Reset()
	defer viper.Reset()

	testKey := "TEST_ENV_VAR_STRING"
	configVar := "test-config-var-string"
	testValue := "test-value"

	// Set environment variable
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	SetDefaultFromEnv(testKey, configVar, String)

	result := viper.GetString(configVar)
	if result != testValue {
		t.Errorf("Expected %s, got %s", testValue, result)
	}
}

func TestSetDefaultFromEnv_Bool(t *testing.T) {
	// Clean up viper state
	viper.Reset()
	defer viper.Reset()

	tests := []struct {
		name      string
		envValue  string
		expected  bool
		shouldSet bool
	}{
		{
			name:      "true_value",
			envValue:  "true",
			expected:  true,
			shouldSet: true,
		},
		{
			name:      "false_value",
			envValue:  "false",
			expected:  false,
			shouldSet: true,
		},
		{
			name:      "1_value",
			envValue:  "1",
			expected:  true,
			shouldSet: true,
		},
		{
			name:      "0_value",
			envValue:  "0",
			expected:  false,
			shouldSet: true,
		},
		{
			name:      "invalid_value",
			envValue:  "invalid",
			expected:  false,
			shouldSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			testKey := "TEST_ENV_VAR_BOOL"
			configVar := "test-config-var-bool"

			os.Setenv(testKey, tt.envValue)
			defer os.Unsetenv(testKey)

			SetDefaultFromEnv(testKey, configVar, Bool)

			if tt.shouldSet {
				result := viper.GetBool(configVar)
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			} else {
				// For invalid values, viper should return the zero value
				result := viper.GetBool(configVar)
				if result != false {
					t.Errorf("Expected false for invalid bool value, got %v", result)
				}
			}
		})
	}
}

func TestSetDefaultFromEnv_Int(t *testing.T) {
	// Clean up viper state
	viper.Reset()
	defer viper.Reset()

	tests := []struct {
		name      string
		envValue  string
		expected  int64
		shouldSet bool
	}{
		{
			name:      "positive_int",
			envValue:  "42",
			expected:  42,
			shouldSet: true,
		},
		{
			name:      "negative_int",
			envValue:  "-10",
			expected:  -10,
			shouldSet: true,
		},
		{
			name:      "zero",
			envValue:  "0",
			expected:  0,
			shouldSet: true,
		},
		{
			name:      "invalid_value",
			envValue:  "not-a-number",
			expected:  0,
			shouldSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			testKey := "TEST_ENV_VAR_INT"
			configVar := "test-config-var-int"

			os.Setenv(testKey, tt.envValue)
			defer os.Unsetenv(testKey)

			SetDefaultFromEnv(testKey, configVar, Int)

			if tt.shouldSet {
				result := viper.GetInt64(configVar)
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			} else {
				// For invalid values, viper should return the zero value
				result := viper.GetInt64(configVar)
				if result != 0 {
					t.Errorf("Expected 0 for invalid int value, got %d", result)
				}
			}
		})
	}
}

func TestSetDefaultFromEnv_MissingEnvVar(t *testing.T) {
	// Clean up viper state
	viper.Reset()
	defer viper.Reset()

	testKey := "NONEXISTENT_ENV_VAR"
	configVar := "test-config-var"

	// Ensure the env var doesn't exist
	os.Unsetenv(testKey)

	// This should not panic or error, just not set anything
	SetDefaultFromEnv(testKey, configVar, String)

	// The config var should not be set
	if viper.IsSet(configVar) {
		t.Error("Config var should not be set when env var doesn't exist")
	}
}

func TestSetDefaultsFromConfig(t *testing.T) {
	// Clean up viper state
	viper.Reset()
	defer viper.Reset()

	configMap := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	SetDefaultsFromConfig(configMap)

	if viper.GetString("key1") != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %s", viper.GetString("key1"))
	}
	if viper.GetInt("key2") != 42 {
		t.Errorf("Expected key2 to be 42, got %d", viper.GetInt("key2"))
	}
	if viper.GetBool("key3") != true {
		t.Errorf("Expected key3 to be true, got %v", viper.GetBool("key3"))
	}
}

func TestTildefyPaths(t *testing.T) {
	// Save original viper state
	viper.Reset()
	defer viper.Reset()

	// Test with a path that doesn't contain tilde
	viper.Set(pconstants.ArgModLocation, "/absolute/path")
	viper.Set(pconstants.ArgInstallDir, "/another/absolute/path")

	err := tildefyPaths()
	if err != nil {
		t.Fatalf("tildefyPaths() returned error: %v", err)
	}

	// Paths without tilde should remain unchanged
	if viper.GetString(pconstants.ArgModLocation) != "/absolute/path" {
		t.Error("Absolute path should remain unchanged")
	}
}

func TestSetConfigFromEnv(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	testKey := "TEST_MULTI_CONFIG_VAR"
	testValue := "test-value"
	configs := []string{"config1", "config2", "config3"}

	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	setConfigFromEnv(testKey, configs, String)

	// All configs should be set to the same value
	for _, config := range configs {
		if viper.GetString(config) != testValue {
			t.Errorf("Expected %s to be set to %s, got %s", config, testValue, viper.GetString(config))
		}
	}
}
