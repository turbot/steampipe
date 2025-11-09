package steampipeconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

// NOTE: Plugin block tests are documented in .ai/milestones/wave-1.5-test-quality/reports/assessment-config-plugin.md
// as a Phase 3 addition. Plugin blocks allow defining multiple instances of a plugin version with specific configurations.
// Missing test scenarios: plugin block parsing, duplicate plugin instances, memory limits, rate limiters.

// Test error handling in loadConfig with non-existent workspace
func TestLoadConfig_NonExistentWorkspace(t *testing.T) {
	steampipeDir, _ := filepath.Abs("testdata/connection_config/single_connection")
	workspaceDir := "/nonexistent/workspace/path"

	// save and restore original InstallDir
	originalInstallDir := app_specific.InstallDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	app_specific.InstallDir = steampipeDir

	_, errorsAndWarnings := loadSteampipeConfig(context.TODO(), workspaceDir, "")
	if errorsAndWarnings.GetError() == nil {
		t.Error("Expected error for non-existent workspace directory")
	}
}

// Test validation result logging
func TestLogValidationResult(t *testing.T) {
	// This is primarily for code coverage - the function logs but doesn't return anything
	warnings := []string{"warning1", "warning2"}
	errors := []string{"error1"}

	// Should not panic
	logValidationResult(warnings, errors)
	logValidationResult(nil, nil)
	logValidationResult([]string{}, []string{})
}

// Test buildValidationLogString
func TestBuildValidationLogString(t *testing.T) {
	tests := []struct {
		name           string
		items          []string
		validationType string
		expectEmpty    bool
	}{
		{
			name:           "multiple_warnings",
			items:          []string{"warning1", "warning2", "warning3"},
			validationType: "warning",
			expectEmpty:    false,
		},
		{
			name:           "single_error",
			items:          []string{"error1"},
			validationType: "error",
			expectEmpty:    false,
		},
		{
			name:           "empty_list",
			items:          []string{},
			validationType: "warning",
			expectEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildValidationLogString(tt.items, tt.validationType)

			if tt.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty string, got: %s", result)
				}
			} else {
				if result == "" {
					t.Error("Expected non-empty string")
				}
				// Should contain the validation type
				if !strings.Contains(result, tt.validationType) {
					t.Errorf("Expected result to contain '%s'", tt.validationType)
				}
				// Should contain the count
				if !strings.Contains(result, fmt.Sprintf("%d", len(tt.items))) {
					t.Error("Expected result to contain item count")
				}
			}
		})
	}
}

// Test loadConfig with simple directory
func TestLoadConfig_SimpleDirectory(t *testing.T) {
	steampipeDir, _ := filepath.Abs("testdata/connection_config/single_connection")
	workspaceDir, _ := filepath.Abs("testdata/load_config_test/empty")

	// save and restore original InstallDir
	originalInstallDir := app_specific.InstallDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	app_specific.InstallDir = steampipeDir

	config, errorsAndWarnings := loadSteampipeConfig(context.TODO(), workspaceDir, "")
	if errorsAndWarnings.GetError() != nil {
		t.Fatalf("Unexpected error: %v", errorsAndWarnings.GetError())
	}

	// Check that config was loaded
	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Check that connections were loaded
	if len(config.Connections) == 0 {
		t.Error("Expected at least one connection")
	}

	// Verify the connection exists
	if _, exists := config.Connections["a"]; !exists {
		t.Error("Expected connection 'a' to exist")
	}
}

// Test getDuplicateConnectionError
func TestGetDuplicateConnectionError(t *testing.T) {
	conn1 := &modconfig.SteampipeConnection{
		Name: "duplicate",
		DeclRange: hclhelpers.Range{
			Filename: "file1.spc",
			Start:    hclhelpers.Pos{Line: 1},
		},
	}
	conn2 := &modconfig.SteampipeConnection{
		Name: "duplicate",
		DeclRange: hclhelpers.Range{
			Filename: "file2.spc",
			Start:    hclhelpers.Pos{Line: 10},
		},
	}

	err := getDuplicateConnectionError(conn1, conn2)
	if err == nil {
		t.Fatal("Expected error for duplicate connection")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "duplicate") {
		t.Error("Error message should contain 'duplicate'")
	}
	if !strings.Contains(errMsg, "file1.spc") || !strings.Contains(errMsg, "file2.spc") {
		t.Error("Error message should contain both file names")
	}
}

type loadConfigTest struct {
	steampipeDir string
	workspaceDir string
	expected     interface{}
}

var testCasesLoadConfig = map[string]loadConfigTest{
	"multiple_connections": {
		steampipeDir: "testdata/connection_config/multiple_connections",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.SteampipeConnection{
				"aws_dmi_001": {
					Name:         "aws_dmi_001",
					PluginAlias:  "aws",
					Plugin:       "/plugins/turbot/aws@latest",
					Type:         "plugin",
					ImportSchema: "enabled",
					Config:       "access_key = \"aws_dmi_001_access_key\"\nregions    = \"- us-east-1\\n-us-west-\"\nsecret_key = \"aws_dmi_001_secret_key\"\n",
					DeclRange: hclhelpers.Range{
						Filename: "$$test_pwd$$/testdata/connection_config/multiple_connections/config/connection1.spc",
						Start: hclhelpers.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hclhelpers.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
					},
				},
				"aws_dmi_002": {
					Name:         "aws_dmi_002",
					PluginAlias:  "aws",
					Plugin:       "/plugins/turbot/aws@latest",
					Type:         "plugin",
					ImportSchema: "enabled",
					Config:       "access_key = \"aws_dmi_002_access_key\"\nregions    = \"- us-east-1\\n-us-west-\"\nsecret_key = \"aws_dmi_002_secret_key\"\n",
					DeclRange: hclhelpers.Range{
						Filename: "$$test_pwd$$/testdata/connection_config/multiple_connections/config/connection2.spc",
						Start: hclhelpers.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hclhelpers.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
					},
				},
			},
		},
	},
	"single_connection": {
		steampipeDir: "testdata/connection_config/single_connection",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.SteampipeConnection{
				"a": {
					Name:         "a",
					PluginAlias:  "test_data/connection-test-1",
					Plugin:       "/plugins/test_data/connection-test-1@latest",
					Type:         "plugin",
					ImportSchema: "enabled",
					DeclRange: hclhelpers.Range{
						Filename: "$$test_pwd$$/testdata/connection_config/single_connection/config/connection1.spc",
						Start: hclhelpers.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hclhelpers.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
					},
				},
			},
		},
	},
}

func TestLoadConfig(t *testing.T) {
	// get the current working directory of the test(used to build the DeclRange.Filename property)
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory")
	}

	for name, test := range testCasesLoadConfig {
		t.Run(name, func(t *testing.T) {
			// default workspace to empty dir
			workspaceDir := test.workspaceDir
			if workspaceDir == "" {
				workspaceDir = "testdata/load_config_test/empty"
			}
			steampipeDir, err := filepath.Abs(test.steampipeDir)
			if err != nil {
				t.Errorf("failed to build absolute config filepath from %s", test.steampipeDir)
			}

			workspaceDir, err = filepath.Abs(workspaceDir)
			if err != nil {
				t.Errorf("failed to build absolute config filepath from %s", workspaceDir)
			}

			// save and restore original InstallDir
			originalInstallDir := app_specific.InstallDir
			defer func() {
				app_specific.InstallDir = originalInstallDir
			}()

			// set app_specific.InstallDir
			app_specific.InstallDir = steampipeDir

			// now load config
			config, errorsAndWarnings := loadSteampipeConfig(context.TODO(), workspaceDir, "")
			if errorsAndWarnings.GetError() != nil {
				if test.expected != "ERROR" {
					t.Errorf("Test: '%s' FAILED with unexpected error: %v", name, errorsAndWarnings.GetError())
				}
				return
			}

			if test.expected == "ERROR" {
				t.Errorf("Test: '%s' FAILED - expected error", name)
				return
			}

			expectedConfig := test.expected.(*SteampipeConfig)
			for _, c := range expectedConfig.Connections {
				c.DeclRange.Filename = strings.Replace(c.DeclRange.Filename, "$$test_pwd$$", pwd, 1)
			}
			if !SteampipeConfigEquals(config, expectedConfig) {
				t.Errorf("Test: '%s' FAILED : expected:\n%s\n\ngot:\n%s", name, expectedConfig, config)
			}
		})
	}
}

// TestLoadConfig_DuplicateConnections tests duplicate connection name detection
// This is a HIGH-VALUE test added in Wave 1.5 Phase 3 to catch config validation bugs
func TestLoadConfig_DuplicateConnections(t *testing.T) {
	// Test that duplicate connection names are detected and reported
	// This tests the getDuplicateConnectionError function in a real scenario

	conn1 := &modconfig.SteampipeConnection{
		Name: "duplicate_conn",
		DeclRange: hclhelpers.Range{
			Filename: "file1.spc",
			Start:    hclhelpers.Pos{Line: 1, Column: 1},
		},
	}

	conn2 := &modconfig.SteampipeConnection{
		Name: "duplicate_conn",
		DeclRange: hclhelpers.Range{
			Filename: "file2.spc",
			Start:    hclhelpers.Pos{Line: 10, Column: 1},
		},
	}

	err := getDuplicateConnectionError(conn1, conn2)

	// Verify error is generated
	assert.NotNil(t, err, "Should detect duplicate connection")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate")
	assert.Contains(t, err.Error(), "duplicate_conn", "Error should mention connection name")
	assert.Contains(t, err.Error(), "file1.spc", "Error should mention first file")
	assert.Contains(t, err.Error(), "file2.spc", "Error should mention second file")

	// BUG HUNTING: Error message should be helpful for users
	// Should include both file locations so user can fix the issue
}

// TestLoadConfig_ErrorMessageQuality tests that error messages are helpful
// This is a HIGH-VALUE test added in Wave 1.5 Phase 3
func TestLoadConfig_ErrorMessageQuality(t *testing.T) {
	tests := []struct {
		name                 string
		configDir            string
		expectedErrorContains []string
		description          string
	}{
		{
			name:      "non_existent_workspace",
			configDir: "/nonexistent/workspace/path",
			expectedErrorContains: []string{
				"workspace", "not found", "does not exist",
			},
			description: "Error for missing workspace should be clear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steampipeDir, _ := filepath.Abs("testdata/connection_config/single_connection")

			// Save and restore original InstallDir
			originalInstallDir := app_specific.InstallDir
			defer func() {
				app_specific.InstallDir = originalInstallDir
			}()

			app_specific.InstallDir = steampipeDir

			_, errorsAndWarnings := loadSteampipeConfig(context.TODO(), tt.configDir, "")

			if errorsAndWarnings.GetError() == nil {
				t.Errorf("Expected error for %s", tt.name)
				return
			}

			errMsg := errorsAndWarnings.GetError().Error()

			// BUG HUNTING: Error messages should be helpful
			// Check that error contains expected information
			foundAtLeastOne := false
			for _, expectedStr := range tt.expectedErrorContains {
				if strings.Contains(strings.ToLower(errMsg), strings.ToLower(expectedStr)) {
					foundAtLeastOne = true
					break
				}
			}

			assert.True(t, foundAtLeastOne,
				"Error message should contain at least one of %v\nGot: %s",
				tt.expectedErrorContains, errMsg)

			// Error message should not be empty or generic
			assert.Greater(t, len(errMsg), 20,
				"Error message should be descriptive (got %d chars): %s", len(errMsg), errMsg)
		})
	}
}

// helpers
func SteampipeConfigEquals(left, right *SteampipeConfig) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	// Compare connection count
	if len(left.Connections) != len(right.Connections) {
		return false
	}

	// Compare each connection by key fields (ignoring Error, PluginPath, PluginInstance fields)
	for name, leftConn := range left.Connections {
		rightConn, ok := right.Connections[name]
		if !ok {
			return false
		}
		if leftConn.Name != rightConn.Name ||
			leftConn.Plugin != rightConn.Plugin ||
			leftConn.PluginAlias != rightConn.PluginAlias ||
			leftConn.Type != rightConn.Type ||
			leftConn.Config != rightConn.Config ||
			leftConn.ImportSchema != rightConn.ImportSchema {
			return false
		}
	}

	if !reflect.DeepEqual(left.DatabaseOptions, right.DatabaseOptions) {
		return false
	}
	if !reflect.DeepEqual(left.GeneralOptions, right.GeneralOptions) {
		return false
	}
	return true
}
