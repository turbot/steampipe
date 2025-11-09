package export

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestGenerateDefaultExportFileName_Format tests the filename format
func TestGenerateDefaultExportFileName_Format(t *testing.T) {
	tests := map[string]struct {
		executionName string
		extension     string
	}{
		"json export": {
			executionName: "query_results",
			extension:     ".json",
		},
		"csv export": {
			executionName: "data_export",
			extension:     ".csv",
		},
		"snapshot export": {
			executionName: "my_snapshot",
			extension:     ".sps",
		},
		"execution name with underscores": {
			executionName: "test_query_name",
			extension:     ".json",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			filename := GenerateDefaultExportFileName(tc.executionName, tc.extension)

			// Verify filename starts with execution name
			assert.True(t, strings.HasPrefix(filename, tc.executionName))

			// Verify filename ends with extension
			assert.True(t, strings.HasSuffix(filename, tc.extension))

			// Verify filename contains a timestamp
			// Format: executionName.YYYYMMDDTHHMMSS.extension
			parts := strings.Split(filename, ".")
			assert.GreaterOrEqual(t, len(parts), 3)

			// The middle part should be the timestamp
			timestampPart := parts[len(parts)-2]
			assert.Len(t, timestampPart, 15, "timestamp should be YYYYMMDDTHHMMSS format (15 chars)")

			// Verify the timestamp contains T separator
			assert.Contains(t, timestampPart, "T")
		})
	}
}

// TestGenerateDefaultExportFileName_Uniqueness tests that filenames are unique
func TestGenerateDefaultExportFileName_Uniqueness(t *testing.T) {
	executionName := "test"
	extension := ".json"

	// Generate two filenames quickly
	filename1 := GenerateDefaultExportFileName(executionName, extension)
	time.Sleep(1 * time.Second)
	filename2 := GenerateDefaultExportFileName(executionName, extension)

	// They should be different due to timestamp
	assert.NotEqual(t, filename1, filename2)
}

// TestGenerateDefaultExportFileName_Components tests individual components
func TestGenerateDefaultExportFileName_Components(t *testing.T) {
	executionName := "test_query"
	extension := ".json"

	filename := GenerateDefaultExportFileName(executionName, extension)

	// Parse the filename components
	// Expected format: test_query.YYYYMMDDTHHMMSS.json
	parts := strings.Split(filename, ".")

	assert.Equal(t, executionName, parts[0], "first part should be execution name")
	assert.True(t, strings.HasPrefix(parts[len(parts)-1], "json"), "last part should be extension")

	// Check timestamp is reasonable (starts with current year)
	now := time.Now()
	currentYear := fmt.Sprintf("%d", now.Year())
	timestampPart := parts[1]
	assert.True(t, strings.HasPrefix(timestampPart, currentYear), "timestamp should start with current year")
}

// TestWrite_InvalidPath tests Write with an invalid file path
func TestWrite_InvalidPath(t *testing.T) {
	// Try to write to a path that doesn't exist
	invalidPath := "/nonexistent/directory/that/does/not/exist/file.json"
	data := strings.NewReader(`{"test": "data"}`)

	err := Write(invalidPath, data)
	assert.Error(t, err, "writing to invalid path should fail")
}

// TestWrite_SuccessfulWrite tests Write with valid path
func TestWrite_SuccessfulWrite(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	filePath := filepath.Join(tempDir, "test.json")
	data := strings.NewReader(`{"test": "data"}`)

	err := Write(filePath, data)
	assert.NoError(t, err)

	// Verify file exists and has correct content
	assert.True(t, helpers.FileExists(filePath))
	content := helpers.ReadTestFile(t, filePath)
	assert.Equal(t, `{"test": "data"}`, content)
}
