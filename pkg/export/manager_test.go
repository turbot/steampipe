package export

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

type testExporter struct {
	alias     string
	extension string
	name      string
}

func (t *testExporter) Export(ctx context.Context, input ExportSourceData, destPath string) error {
	return nil
}
func (t *testExporter) FileExtension() string { return t.extension }
func (t *testExporter) Name() string          { return t.name }
func (t *testExporter) Alias() string         { return t.alias }

var dummyCSVExporter = testExporter{alias: "", extension: ".csv", name: "csv"}
var dummyJSONExporter = testExporter{alias: "", extension: ".json", name: "json"}
var dummyASFFExporter = testExporter{alias: "asff.json", extension: ".json", name: "asff"}
var dummyNUNITExporter = testExporter{alias: "nunit3.xml", extension: ".xml", name: "nunit3"}
var dummySPSExporter = testExporter{alias: "sps", extension: constants.SnapshotExtension, name: constants.OutputFormatSnapshot}

// testExportSourceData is a simple implementation of ExportSourceData for testing
type testExportSourceData struct{}

func (t *testExportSourceData) IsExportSourceData() {}

type exporterTestCase struct {
	name   string
	input  string
	expect interface{}
}

// TestManagerRegistration tests exporter registration
func TestManagerRegistration(t *testing.T) {
	tests := map[string]struct {
		registerFirst  *testExporter
		registerSecond *testExporter
		expectError    bool
		description    string
	}{
		"register single exporter": {
			registerFirst: &dummyJSONExporter,
			expectError:   false,
			description:   "should successfully register a single exporter",
		},
		"register duplicate name": {
			registerFirst:  &dummyJSONExporter,
			registerSecond: &dummyJSONExporter,
			expectError:    true,
			description:    "should fail when registering exporter with duplicate name",
		},
		"register different exporters": {
			registerFirst:  &dummyJSONExporter,
			registerSecond: &dummyCSVExporter,
			expectError:    false,
			description:    "should successfully register different exporters",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewManager()
			err := m.Register(tc.registerFirst)
			assert.NoError(t, err, "first registration should succeed")

			if tc.registerSecond != nil {
				err = m.Register(tc.registerSecond)
				if tc.expectError {
					assert.Error(t, err, tc.description)
				} else {
					assert.NoError(t, err, tc.description)
				}
			}
		})
	}
}

// TestManagerRegisterByAlias tests that exporters can be resolved by alias
func TestManagerRegisterByAlias(t *testing.T) {
	m := NewManager()
	err := m.Register(&dummySPSExporter)
	assert.NoError(t, err)

	// Should be able to resolve by name
	target, err := m.getExportTarget(constants.OutputFormatSnapshot, "test")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, constants.OutputFormatSnapshot, target.exporter.Name())

	// Should also be able to resolve by alias
	target, err = m.getExportTarget("sps", "test")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, constants.OutputFormatSnapshot, target.exporter.Name())
}

// TestManagerRegisterByExtension tests that exporters can be resolved by file extension
func TestManagerRegisterByExtension(t *testing.T) {
	tests := map[string]struct {
		exporter        *testExporter
		testInput       string
		expectError     bool
		expectNamed     bool
		expectedName    string
		expectedExtends string
	}{
		"json by extension": {
			exporter:        &dummyJSONExporter,
			testInput:       "output.json",
			expectError:     false,
			expectNamed:     true,
			expectedName:    "json",
			expectedExtends: ".json",
		},
		"csv by extension": {
			exporter:        &dummyCSVExporter,
			testInput:       "output.csv",
			expectError:     false,
			expectNamed:     true,
			expectedName:    "csv",
			expectedExtends: ".csv",
		},
		"snapshot by extension": {
			exporter:        &dummySPSExporter,
			testInput:       "output.sps",
			expectError:     false,
			expectNamed:     true,
			expectedName:    constants.OutputFormatSnapshot,
			expectedExtends: constants.SnapshotExtension,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewManager()
			err := m.Register(tc.exporter)
			assert.NoError(t, err)

			target, err := m.getExportTarget(tc.testInput, "test")
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, target)
				assert.Equal(t, tc.expectedName, target.exporter.Name())
				assert.Equal(t, tc.expectNamed, target.isNamedTarget)
			}
		})
	}
}

// TestValidateExportFormat tests export format validation
func TestValidateExportFormat(t *testing.T) {
	m := NewManager()
	m.Register(&dummyJSONExporter)
	m.Register(&dummyCSVExporter)
	m.Register(&dummySPSExporter)

	tests := map[string]struct {
		exports     []string
		expectError bool
		description string
	}{
		"valid single format": {
			exports:     []string{"json"},
			expectError: false,
			description: "single valid format should pass",
		},
		"valid multiple formats": {
			exports:     []string{"json", "csv"},
			expectError: false,
			description: "multiple valid formats should pass",
		},
		"invalid format": {
			exports:     []string{"invalid-format"},
			expectError: true,
			description: "invalid format should fail",
		},
		"mix of valid and invalid": {
			exports:     []string{"json", "invalid-format"},
			expectError: true,
			description: "mix of valid and invalid should fail",
		},
		"named export": {
			exports:     []string{"output.json"},
			expectError: false,
			description: "named export should pass",
		},
		"mix named and unnamed": {
			exports:     []string{"json", "output.csv"},
			expectError: true,
			description: "mixing named and unnamed exports should fail",
		},
		"empty exports": {
			exports:     []string{},
			expectError: false,
			description: "empty exports should pass",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := m.ValidateExportFormat(tc.exports)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestResolveTargetsFromArgs tests target resolution from export arguments
func TestResolveTargetsFromArgs(t *testing.T) {
	m := NewManager()
	m.Register(&dummyJSONExporter)
	m.Register(&dummyCSVExporter)

	tests := map[string]struct {
		args          []string
		executionName string
		expectCount   int
		expectError   bool
	}{
		"single format": {
			args:          []string{"json"},
			executionName: "test_query",
			expectCount:   1,
			expectError:   false,
		},
		"multiple formats": {
			args:          []string{"json", "csv"},
			executionName: "test_query",
			expectCount:   2,
			expectError:   false,
		},
		"duplicate formats": {
			args:          []string{"json", "json"},
			executionName: "test_query",
			expectCount:   1, // Should deduplicate
			expectError:   false,
		},
		"named exports": {
			args:          []string{"output.json", "data.csv"},
			executionName: "test_query",
			expectCount:   2,
			expectError:   false,
		},
		"invalid format": {
			args:          []string{"invalid"},
			executionName: "test_query",
			expectCount:   0,
			expectError:   true,
		},
		"empty string": {
			args:          []string{"", "json"},
			executionName: "test_query",
			expectCount:   1, // Empty strings should be ignored
			expectError:   false,
		},
		"whitespace only": {
			args:          []string{"  ", "json"},
			executionName: "test_query",
			expectCount:   1, // Whitespace-only strings should be ignored
			expectError:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			targets, err := m.resolveTargetsFromArgs(tc.args, tc.executionName)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, targets, tc.expectCount)
			}
		})
	}
}

// fileWritingExporter is a test exporter that writes to files
type fileWritingExporter struct {
	name      string
	extension string
}

func (f *fileWritingExporter) Export(ctx context.Context, input ExportSourceData, destPath string) error {
	return Write(destPath, strings.NewReader(`{"test": "data"}`))
}

func (f *fileWritingExporter) FileExtension() string {
	return f.extension
}

func (f *fileWritingExporter) Name() string {
	return f.name
}

func (f *fileWritingExporter) Alias() string {
	return ""
}

// TestDoExportWithRealFiles tests the complete export flow with file creation
func TestDoExportWithRealFiles(t *testing.T) {
	jsonExporter := &fileWritingExporter{
		name:      "json",
		extension: ".json",
	}

	m := NewManager()
	err := m.Register(jsonExporter)
	assert.NoError(t, err)

	tempDir := helpers.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "output.json")

	// Create test data
	testData := &testExportSourceData{}

	// Export to named file
	locations, err := m.DoExport(context.Background(), "test", testData, []string{outputPath})
	assert.NoError(t, err)
	assert.Len(t, locations, 1)

	// Verify file was created
	assert.True(t, helpers.FileExists(outputPath))

	// Verify file content
	content := helpers.ReadTestFile(t, outputPath)
	assert.Equal(t, `{"test": "data"}`, content)
}

// TestGenerateDefaultExportFileName tests the default filename generation
func TestGenerateDefaultExportFileName(t *testing.T) {
	executionName := "test_query"
	extension := ".json"

	filename := GenerateDefaultExportFileName(executionName, extension)

	// Check that filename contains execution name
	assert.Contains(t, filename, executionName)

	// Check that filename ends with the extension
	assert.True(t, strings.HasSuffix(filename, extension))

	// Check that filename contains a timestamp (has the pattern YYYYMMDDTHHMMSS)
	// The format should be: test_query.YYYYMMDDTHHMMSS.json
	parts := strings.Split(filename, ".")
	assert.GreaterOrEqual(t, len(parts), 3) // name, timestamp, extension
}

