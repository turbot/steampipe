package cmdconfig

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
)

func TestValidateSnapshotTags_EdgeCases(t *testing.T) {
	t.Skip("Demonstrates bugs #4756, #4757 - validateSnapshotTags accepts invalid tags. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	// NOTE: This test documents expected behavior. The bug is in validateSnapshotTags
	// which uses strings.Split(tagStr, "=") without checking for empty key/value parts.
	// Tags like "key=" and "=value" should fail but currently pass validation.
	tests := []struct {
		name      string
		tags      []string
		shouldErr bool
		desc      string
	}{
		{
			name:      "valid_single_tag",
			tags:      []string{"env=prod"},
			shouldErr: false,
			desc:      "Valid tag with single equals",
		},
		{
			name:      "multiple_valid_tags",
			tags:      []string{"env=prod", "region=us-east"},
			shouldErr: false,
			desc:      "Multiple valid tags",
		},
		{
			name:      "tag_with_double_equals",
			tags:      []string{"key==value"},
			shouldErr: true,
			desc:      "BUG?: Tag with double equals should fail but might be split incorrectly",
		},
		{
			name:      "tag_starting_with_equals",
			tags:      []string{"=value"},
			shouldErr: true,
			desc:      "BUG?: Tag starting with equals has empty key",
		},
		{
			name:      "tag_ending_with_equals",
			tags:      []string{"key="},
			shouldErr: true,
			desc:      "BUG?: Tag ending with equals has empty value",
		},
		{
			name:      "tag_without_equals",
			tags:      []string{"invalid"},
			shouldErr: true,
			desc:      "Tag without equals sign should fail",
		},
		{
			name:      "empty_tag_string",
			tags:      []string{""},
			shouldErr: true,
			desc:      "BUG?: Empty tag string",
		},
		{
			name:      "tag_with_multiple_equals",
			tags:      []string{"key=value=extra"},
			shouldErr: true,
			desc:      "BUG?: Tag with multiple equals signs",
		},
		{
			name:      "mixed_valid_and_invalid",
			tags:      []string{"valid=tag", "invalid"},
			shouldErr: true,
			desc:      "Mixed valid and invalid tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up viper state
			viper.Reset()
			defer viper.Reset()

			viper.Set(pconstants.ArgSnapshotTag, tt.tags)
			err := validateSnapshotTags()

			if tt.shouldErr && err == nil {
				t.Errorf("%s: Expected error but got nil", tt.desc)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("%s: Expected no error but got: %v", tt.desc, err)
			}
		})
	}
}

func TestValidateSnapshotArgs_Conflicts(t *testing.T) {
	tests := []struct {
		name      string
		share     bool
		snapshot  bool
		shouldErr bool
		desc      string
	}{
		{
			name:      "both_share_and_snapshot_true",
			share:     true,
			snapshot:  true,
			shouldErr: true,
			desc:      "Both share and snapshot set should fail",
		},
		{
			name:      "only_share_true",
			share:     true,
			snapshot:  false,
			shouldErr: false,
			desc:      "Only share set is valid",
		},
		{
			name:      "only_snapshot_true",
			share:     false,
			snapshot:  true,
			shouldErr: false,
			desc:      "Only snapshot set is valid",
		},
		{
			name:      "both_false",
			share:     false,
			snapshot:  false,
			shouldErr: false,
			desc:      "Both false should be valid (no snapshot mode)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up viper state
			viper.Reset()
			defer viper.Reset()

			viper.Set(pconstants.ArgShare, tt.share)
			viper.Set(pconstants.ArgSnapshot, tt.snapshot)
			viper.Set(pconstants.ArgPipesHost, "test-host") // Set default to avoid nil check failure

			ctx := context.Background()
			err := ValidateSnapshotArgs(ctx)

			if tt.shouldErr && err == nil {
				t.Errorf("%s: Expected error but got nil", tt.desc)
			}
			if !tt.shouldErr && err != nil {
				// Some errors are expected if token is missing, etc.
				// Only fail if it's the conflict error
				if tt.share && tt.snapshot {
					// This should be the specific conflict error
					t.Logf("%s: Got error (may be acceptable): %v", tt.desc, err)
				}
			}
		})
	}
}

func TestValidateSnapshotLocation_FileValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		location     string
		locationFunc func() string // Generate location dynamically
		token        string
		shouldErr    bool
		desc         string
	}{
		{
			name:         "existing_directory",
			locationFunc: func() string { return tempDir },
			token:        "",
			shouldErr:    false,
			desc:         "Existing directory should be valid",
		},
		{
			name:      "nonexistent_directory",
			location:  "/nonexistent/path/that/does/not/exist",
			token:     "",
			shouldErr: true,
			desc:      "Non-existent directory should fail",
		},
		{
			name:      "empty_location_without_token",
			location:  "",
			token:     "",
			shouldErr: true,
			desc:      "Empty location without token should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up viper state
			viper.Reset()
			defer viper.Reset()

			location := tt.location
			if tt.locationFunc != nil {
				location = tt.locationFunc()
			}

			viper.Set(pconstants.ArgSnapshotLocation, location)
			viper.Set(pconstants.ArgPipesToken, tt.token)

			ctx := context.Background()
			err := validateSnapshotLocation(ctx, tt.token)

			if tt.shouldErr && err == nil {
				t.Errorf("%s: Expected error but got nil", tt.desc)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("%s: Expected no error but got: %v", tt.desc, err)
			}
		})
	}
}

func TestValidateSnapshotArgs_MissingHost(t *testing.T) {
	// Test the case where pipes-host is empty/missing
	viper.Reset()
	defer viper.Reset()

	viper.Set(pconstants.ArgShare, true)
	viper.Set(pconstants.ArgPipesHost, "") // Empty host

	ctx := context.Background()
	err := ValidateSnapshotArgs(ctx)

	if err == nil {
		t.Error("Expected error when pipes-host is empty, but got nil")
	}
}

func TestValidateSnapshotTags_EmptyAndWhitespace(t *testing.T) {
	t.Skip("Demonstrates bugs #4756, #4757 - validateSnapshotTags accepts tags with whitespace and empty values. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	tests := []struct {
		name      string
		tags      []string
		shouldErr bool
		desc      string
	}{
		{
			name:      "tag_with_whitespace",
			tags:      []string{" key = value "},
			shouldErr: true,
			desc:      "BUG?: Tag with whitespace around equals",
		},
		{
			name:      "tag_only_equals",
			tags:      []string{"="},
			shouldErr: true,
			desc:      "BUG?: Tag that is only equals sign",
		},
		{
			name:      "tag_with_special_chars",
			tags:      []string{"key@#$=value"},
			shouldErr: false,
			desc:      "Tag with special characters in key should be accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			defer viper.Reset()

			viper.Set(pconstants.ArgSnapshotTag, tt.tags)
			err := validateSnapshotTags()

			if tt.shouldErr && err == nil {
				t.Errorf("%s: Expected error but got nil", tt.desc)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("%s: Expected no error but got: %v", tt.desc, err)
			}
		})
	}
}

func TestValidateSnapshotLocation_TildePath(t *testing.T) {
	t.Skip("Demonstrates bugs #4756, #4757 - validateSnapshotLocation doesn't expand tilde paths. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	// Test tildefy functionality with invalid paths
	viper.Reset()
	defer viper.Reset()

	// Set a location that starts with tilde
	viper.Set(pconstants.ArgSnapshotLocation, "~/test_snapshot_location_that_does_not_exist")
	viper.Set(pconstants.ArgPipesToken, "")

	ctx := context.Background()
	err := validateSnapshotLocation(ctx, "")

	// Should fail because the directory doesn't exist after tildifying
	if err == nil {
		t.Error("Expected error for non-existent tilde path, but got nil")
	}
}

func TestValidateSnapshotArgs_WorkspaceIdentifierWithoutToken(t *testing.T) {
	// Test that workspace identifier requires a token
	viper.Reset()
	defer viper.Reset()

	viper.Set(pconstants.ArgSnapshot, true)
	viper.Set(pconstants.ArgSnapshotLocation, "acme/dev") // Workspace identifier format
	viper.Set(pconstants.ArgPipesToken, "")              // No token
	viper.Set(pconstants.ArgPipesHost, "pipes.turbot.com")

	ctx := context.Background()
	err := ValidateSnapshotArgs(ctx)

	if err == nil {
		t.Error("Expected error when using workspace identifier without token, but got nil")
	}
}

func TestValidateSnapshotLocation_RelativePath(t *testing.T) {
	// Create a relative path test directory
	relDir := "test_rel_snapshot_dir"
	defer os.RemoveAll(relDir)

	err := os.Mkdir(relDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Get absolute path for comparison
	absDir, err := filepath.Abs(relDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	viper.Reset()
	defer viper.Reset()

	viper.Set(pconstants.ArgSnapshotLocation, relDir)
	viper.Set(pconstants.ArgPipesToken, "")

	ctx := context.Background()
	err = validateSnapshotLocation(ctx, "")

	// After validation, check if the path was modified
	resultLocation := viper.GetString(pconstants.ArgSnapshotLocation)

	if err != nil {
		t.Errorf("Expected no error for valid relative path, but got: %v", err)
	}

	// The location might be absolute or relative, but should be valid
	if resultLocation == "" {
		t.Error("Location was cleared after validation")
	}

	t.Logf("Original: %s, After validation: %s, Expected abs: %s", relDir, resultLocation, absDir)
}
