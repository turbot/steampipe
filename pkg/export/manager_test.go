package export

import (
	"context"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/constants"
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

type exporterTestCase struct {
	name   string
	input  string
	expect interface{}
}

var exporterTestCases = []exporterTestCase{
	{
		name:   "Bad Format",
		input:  "bad-format",
		expect: "ERROR",
	},
	{
		name:   "csv file name",
		input:  "file.csv",
		expect: &dummyCSVExporter,
	},
	{
		name:   "csv format name",
		input:  "csv",
		expect: &dummyCSVExporter,
	},
	{
		name:   "Snapshot file name",
		input:  "file.sps",
		expect: &dummySPSExporter,
	},
	{
		name:   "Snapshot format name",
		input:  "sps",
		expect: &dummySPSExporter,
	},
	{
		name:   "json file name",
		input:  "file.json",
		expect: &dummyJSONExporter,
	},
	{
		name:   "json format name",
		input:  "json",
		expect: &dummyJSONExporter,
	},
	{
		name:   "asff json file name",
		input:  "file.asff.json",
		expect: &dummyASFFExporter,
	},
	{
		name:   "asff json format name",
		input:  "asff.json",
		expect: &dummyASFFExporter,
	},
	{
		name:   "nunit3 file name",
		input:  "file.nunit3.xml",
		expect: &dummyNUNITExporter,
	},
	{
		name:   "nunit3 format name",
		input:  "nunit3.xml",
		expect: &dummyNUNITExporter,
	},
}

func TestDoExport(t *testing.T) {
	exportersToRegister := []*testExporter{
		&dummyJSONExporter,
		&dummyCSVExporter,
		&dummySPSExporter,
		&dummyASFFExporter,
		&dummyNUNITExporter,
	}

	m := NewManager()
	for _, e := range exportersToRegister {
		m.Register(e)
	}
	for _, testCase := range exporterTestCases {
		targets, err := m.resolveTargetsFromArgs([]string{testCase.input}, "dummy_execution_name")
		shouldError := testCase.expect == "ERROR"
		if shouldError {
			if err == nil {
				t.Errorf("Request for '%s' should have errored - but did not", testCase.input)
			}
			continue
		}
		if !shouldError {
			if err != nil {
				t.Errorf("Request for '%s' should not have errored - but did: %v", testCase.input, err)
			}
			continue
		}

		if len(targets) != 1 {
			t.Errorf("%v with %v input => expected one target - got %d", testCase.name, testCase.input, len(targets))
			continue
		}
		actualTarget := targets[0]
		expectedTargetExporter := testCase.expect.(*testExporter)

		if actualTarget.exporter != expectedTargetExporter {
			t.Errorf("%v with %v input => expected %s target - got %s", testCase.name, testCase.input, testCase.expect.(*testExporter).Name(), actualTarget.exporter.Name())
			continue
		}
	}
}

// TestManager_ConcurrentRegistration tests that the Manager can handle concurrent
// exporter registration safely. This test is designed to expose race conditions
// when run with the -race flag.
//
// Related issue: #4715
func TestManager_ConcurrentRegistration(t *testing.T) {
	// Create a manager instance
	m := NewManager()

	// Create multiple test exporters with unique names
	exporters := []*testExporter{
		{alias: "", extension: ".csv", name: "csv"},
		{alias: "", extension: ".json", name: "json"},
		{alias: "", extension: ".xml", name: "xml"},
		{alias: "", extension: ".html", name: "html"},
		{alias: "", extension: ".yaml", name: "yaml"},
		{alias: "", extension: ".md", name: "markdown"},
		{alias: "", extension: ".txt", name: "text"},
		{alias: "", extension: ".log", name: "log"},
	}

	// Channel to collect errors from goroutines
	errChan := make(chan error, len(exporters))
	done := make(chan bool)

	// Register all exporters concurrently
	for _, exp := range exporters {
		go func(e *testExporter) {
			err := m.Register(e)
			errChan <- err
		}(exp)
	}

	// Collect results
	go func() {
		for i := 0; i < len(exporters); i++ {
			err := <-errChan
			if err != nil {
				t.Errorf("Failed to register exporter: %v", err)
			}
		}
		done <- true
	}()

	// Wait for completion
	<-done

	// Verify all exporters were registered successfully
	// Each exporter should be accessible by its name
	for _, exp := range exporters {
		target, err := m.getExportTarget(exp.name, "test_exec")
		if err != nil {
			t.Errorf("Exporter '%s' was not registered properly: %v", exp.name, err)
		}
		if target == nil {
			t.Errorf("Exporter '%s' returned nil target", exp.name)
		}
	}
}
