package export

import (
	"context"
	"testing"

	"github.com/turbot/steampipe/pkg/constants"
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
