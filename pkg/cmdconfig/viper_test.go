package cmdconfig

import (
	"fmt"
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

// Concurrency and race condition tests

func TestViperGlobalState_ConcurrentReads(t *testing.T) {
	// Test concurrent reads from viper - should be safe
	viper.Reset()
	defer viper.Reset()

	viper.Set("test-key", "test-value")

	done := make(chan bool)
	errors := make(chan string, 100)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				val := viper.GetString("test-key")
				if val != "test-value" {
					errors <- fmt.Sprintf("Goroutine %d: Expected 'test-value', got '%s'", id, val)
				}
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestViperGlobalState_ConcurrentWrites(t *testing.T) {
	// t.Skip("Demonstrates bugs #4756, #4757 - Viper global state has race conditions on concurrent writes. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	// Test concurrent writes to viper with mutex protection
	viperMutex.Lock()
	viper.Reset()
	viperMutex.Unlock()
	defer func() {
		viperMutex.Lock()
		viper.Reset()
		viperMutex.Unlock()
	}()

	done := make(chan bool)
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 50; j++ {
				viperMutex.Lock()
				viper.Set("concurrent-key", id)
				viperMutex.Unlock()
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// The final value is now deterministic with mutex protection
	viperMutex.RLock()
	finalVal := viper.GetInt("concurrent-key")
	viperMutex.RUnlock()
	t.Logf("Final value after concurrent writes: %d", finalVal)
}

func TestViperGlobalState_ConcurrentReadWrite(t *testing.T) {
	// t.Skip("Demonstrates bugs #4756, #4757 - Viper global state has race conditions on concurrent read/write. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	// Test concurrent reads and writes with mutex protection
	viperMutex.Lock()
	viper.Reset()
	viper.Set("race-key", "initial")
	viperMutex.Unlock()
	defer func() {
		viperMutex.Lock()
		viper.Reset()
		viperMutex.Unlock()
	}()

	done := make(chan bool)
	numReaders := 5
	numWriters := 5

	// Start readers
	for i := 0; i < numReaders; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				viperMutex.RLock()
				_ = viper.GetString("race-key")
				viperMutex.RUnlock()
			}
		}(i)
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 50; j++ {
				viperMutex.Lock()
				viper.Set("race-key", id)
				viperMutex.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numReaders+numWriters; i++ {
		<-done
	}

	t.Log("Concurrent read/write completed successfully with mutex protection")
}

func TestSetDefaultFromEnv_ConcurrentAccess(t *testing.T) {
	// t.Skip("Demonstrates bugs #4756, #4757 - SetDefaultFromEnv has race conditions on concurrent access. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	// BUG?: Test concurrent access to SetDefaultFromEnv
	viper.Reset()
	defer viper.Reset()

	// Set up multiple env vars
	envVars := make(map[string]string)
	for i := 0; i < 10; i++ {
		key := "TEST_CONCURRENT_ENV_" + string(rune('A'+i))
		val := "value" + string(rune('0'+i))
		envVars[key] = val
		os.Setenv(key, val)
		defer os.Unsetenv(key)
	}

	done := make(chan bool)
	numGoroutines := 10

	// Concurrently set defaults from env
	i := 0
	for key := range envVars {
		go func(envKey string, configVar string) {
			defer func() { done <- true }()
			SetDefaultFromEnv(envKey, configVar, String)
		}(key, "config-var-"+string(rune('A'+i)))
		i++
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Log("Concurrent SetDefaultFromEnv completed")
}

func TestSetDefaultsFromConfig_ConcurrentCalls(t *testing.T) {
	// t.Skip("Demonstrates bugs #4756, #4757 - SetDefaultsFromConfig has race conditions on concurrent calls. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	// BUG?: Test concurrent calls to SetDefaultsFromConfig
	viper.Reset()
	defer viper.Reset()

	done := make(chan bool)
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			configMap := map[string]interface{}{
				"key-" + string(rune('A'+id)): "value-" + string(rune('0'+id)),
			}
			SetDefaultsFromConfig(configMap)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Log("Concurrent SetDefaultsFromConfig completed")
}

func TestSetBaseDefaults_MultipleCalls(t *testing.T) {
	// Test calling setBaseDefaults multiple times
	viper.Reset()
	defer viper.Reset()

	err := setBaseDefaults()
	if err != nil {
		t.Fatalf("First call to setBaseDefaults failed: %v", err)
	}

	// Call again - should be idempotent
	err = setBaseDefaults()
	if err != nil {
		t.Fatalf("Second call to setBaseDefaults failed: %v", err)
	}

	// Verify values are still correct
	if viper.GetString(pconstants.ArgTelemetry) != constants.TelemetryInfo {
		t.Error("Telemetry default changed after second call")
	}
}

func TestViperReset_StateCleanup(t *testing.T) {
	// Test that viper.Reset() properly cleans up state
	viper.Reset()
	defer viper.Reset()

	// Set some values
	viper.Set("test-key-1", "value1")
	viper.Set("test-key-2", 42)
	viper.Set("test-key-3", true)

	// Verify values are set
	if viper.GetString("test-key-1") != "value1" {
		t.Error("Value not set correctly")
	}

	// Reset viper
	viper.Reset()

	// Verify values are cleared
	if viper.GetString("test-key-1") != "" {
		t.Error("BUG?: Viper.Reset() did not clear string value")
	}
	if viper.GetInt("test-key-2") != 0 {
		t.Error("BUG?: Viper.Reset() did not clear int value")
	}
	if viper.GetBool("test-key-3") != false {
		t.Error("BUG?: Viper.Reset() did not clear bool value")
	}
}

func TestSetDefaultFromEnv_TypeConversionErrors(t *testing.T) {
	// Test that type conversion errors are handled gracefully
	viper.Reset()
	defer viper.Reset()

	tests := []struct {
		name      string
		envValue  string
		varType   EnvVarType
		configVar string
		desc      string
	}{
		{
			name:      "invalid_bool",
			envValue:  "not-a-bool",
			varType:   Bool,
			configVar: "test-invalid-bool",
			desc:      "Invalid bool value should not panic",
		},
		{
			name:      "invalid_int",
			envValue:  "not-a-number",
			varType:   Int,
			configVar: "test-invalid-int",
			desc:      "Invalid int value should not panic",
		},
		{
			name:      "empty_string_as_bool",
			envValue:  "",
			varType:   Bool,
			configVar: "test-empty-bool",
			desc:      "Empty string as bool should not panic",
		},
		{
			name:      "empty_string_as_int",
			envValue:  "",
			varType:   Int,
			configVar: "test-empty-int",
			desc:      "Empty string as int should not panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testKey := "TEST_TYPE_CONVERSION_" + tt.name
			os.Setenv(testKey, tt.envValue)
			defer os.Unsetenv(testKey)

			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: Panicked with: %v", tt.desc, r)
				}
			}()

			SetDefaultFromEnv(testKey, tt.configVar, tt.varType)

			t.Logf("%s: Handled gracefully", tt.desc)
		})
	}
}

func TestTildefyPaths_InvalidPaths(t *testing.T) {
	// Test tildefyPaths with various invalid paths
	viper.Reset()
	defer viper.Reset()

	tests := []struct {
		name      string
		modLoc    string
		installDir string
		shouldErr bool
		desc      string
	}{
		{
			name:       "empty_paths",
			modLoc:     "",
			installDir: "",
			shouldErr:  false,
			desc:       "Empty paths should be handled gracefully",
		},
		{
			name:       "valid_absolute_paths",
			modLoc:     "/tmp/test",
			installDir: "/tmp/install",
			shouldErr:  false,
			desc:       "Valid absolute paths should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(pconstants.ArgModLocation, tt.modLoc)
			viper.Set(pconstants.ArgInstallDir, tt.installDir)

			err := tildefyPaths()

			if tt.shouldErr && err == nil {
				t.Errorf("%s: Expected error but got nil", tt.desc)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("%s: Expected no error but got: %v", tt.desc, err)
			}
		})
	}
}
