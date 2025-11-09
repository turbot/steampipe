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

// TestRegisterExporterByExtension_ExtensionClash tests the complex logic for handling extension conflicts
func TestRegisterExporterByExtension_ExtensionClash(t *testing.T) {
	tests := map[string]struct {
		first           *testExporter
		second          *testExporter
		expectedWinner  string
		description     string
	}{
		"default exporter wins over non-default": {
			first:          &testExporter{name: "custom", extension: ".json", alias: ""},
			second:         &testExporter{name: "json", extension: ".json", alias: ""},
			expectedWinner: "json",
			description:    "json exporter is default for .json extension",
		},
		"first default keeps precedence": {
			first:          &testExporter{name: "json", extension: ".json", alias: ""},
			second:         &testExporter{name: "custom", extension: ".json", alias: ""},
			expectedWinner: "json",
			description:    "first registered default keeps precedence",
		},
		"neither default registers second": {
			first:          &testExporter{name: "asff", extension: ".json", alias: "asff.json"},
			second:         &testExporter{name: "nunit3", extension: ".json", alias: "nunit3.json"},
			expectedWinner: "nunit3",
			description:    "when neither is default, the second exporter wins",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewManager()

			// Register first exporter
			err := m.Register(tc.first)
			assert.NoError(t, err)

			// Register second exporter
			err = m.Register(tc.second)
			assert.NoError(t, err)

			// Check which exporter is mapped to the extension
			exporter, exists := m.registeredExtensions[".json"]
			assert.True(t, exists, tc.description)
			assert.Equal(t, tc.expectedWinner, exporter.Name(), tc.description)
		})
	}
}

// TestRegisterExporterByExtension_MultiSegmentExtension tests handling of extensions like .asff.json
func TestRegisterExporterByExtension_MultiSegmentExtension(t *testing.T) {
	m := NewManager()

	// Register exporter with multi-segment extension
	asffExporter := &testExporter{
		name:      "asff",
		extension: ".asff.json",
		alias:     "asff.json",
	}

	err := m.Register(asffExporter)
	assert.NoError(t, err)

	// Should be registered under both full extension and short extension
	fullExtExporter, fullExists := m.registeredExtensions[".asff.json"]
	assert.True(t, fullExists)
	assert.Equal(t, "asff", fullExtExporter.Name())

	shortExtExporter, shortExists := m.registeredExtensions[".json"]
	assert.True(t, shortExists)
	assert.Equal(t, "asff", shortExtExporter.Name())
}

// TestHasNamedExport tests detection of named exports
func TestHasNamedExport(t *testing.T) {
	m := NewManager()
	m.Register(&dummyJSONExporter)
	m.Register(&dummyCSVExporter)

	tests := map[string]struct {
		exports  []string
		expected bool
	}{
		"only format names": {
			exports:  []string{"json", "csv"},
			expected: false,
		},
		"only named files": {
			exports:  []string{"output.json", "data.csv"},
			expected: true,
		},
		"mixed named and unnamed": {
			exports:  []string{"json", "output.csv"},
			expected: true, // Has at least one named export
		},
		"single named export": {
			exports:  []string{"report.json"},
			expected: true,
		},
		"single format name": {
			exports:  []string{"json"},
			expected: false,
		},
		"empty list": {
			exports:  []string{},
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := m.HasNamedExport(tc.exports)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDoExport_WithErrors tests error handling in DoExport
func TestDoExport_WithErrors(t *testing.T) {
	tests := map[string]struct {
		exports      []string
		expectError  bool
		expectCount  int
		description  string
	}{
		"empty exports": {
			exports:      []string{},
			expectError:  false,
			expectCount:  0,
			description:  "empty exports should return nil without error",
		},
		"invalid format": {
			exports:      []string{"invalid-format"},
			expectError:  true,
			expectCount:  0,
			description:  "invalid format should return error",
		},
		"partial success with invalid": {
			exports:      []string{"json", "invalid-format"},
			expectError:  true,
			expectCount:  0,
			description:  "should fail if any format is invalid",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewManager()
			m.Register(&fileWritingExporter{name: "json", extension: ".json"})

			testData := &testExportSourceData{}
			locations, err := m.DoExport(context.Background(), "test", testData, tc.exports)

			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
			assert.Len(t, locations, tc.expectCount)
		})
	}
}

// TestRegister_ErrorCases tests error handling in Register
func TestRegister_ErrorCases(t *testing.T) {
	tests := map[string]struct {
		setup       func(*Manager)
		exporter    *testExporter
		expectError bool
		description string
	}{
		"duplicate alias fails": {
			setup: func(m *Manager) {
				// Register exporter with alias "sps"
				m.Register(&testExporter{name: "snapshot1", extension: ".sps", alias: "sps"})
			},
			exporter:    &testExporter{name: "snapshot2", extension: ".snap", alias: "sps"},
			expectError: true,
			description: "registering exporter with duplicate alias should fail",
		},
		"exporter with empty alias succeeds": {
			setup:       func(m *Manager) {},
			exporter:    &testExporter{name: "test", extension: ".test", alias: ""},
			expectError: false,
			description: "exporter with empty alias should succeed",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewManager()
			tc.setup(m)

			err := m.Register(tc.exporter)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

