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

var dummyCSVExporter = testExporter{alias: "", extension: ".csv", name: "CSV"}
var dummyJSONExporter = testExporter{alias: "", extension: ".json", name: "JSON"}
var dummyASFFExporter = testExporter{alias: "asff.json", extension: ".json", name: "ASFF"}
var dummyNUNITExporter = testExporter{alias: "nunit3.xml", extension: ".xml", name: "NUNIT3"}
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
		name:   "csv",
		input:  "file.csv",
		expect: &dummyCSVExporter,
	},
	{
		name:   "Snapshot",
		input:  "file.sps",
		expect: &dummySPSExporter,
	},
	{
		name:   "json",
		input:  "file.json",
		expect: &dummyJSONExporter,
	},
	{
		name:   "asff json",
		input:  "file.asff.json",
		expect: &dummyASFFExporter,
	},
	{
		name:   "nunit3",
		input:  "file.nunit3.xml",
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
				t.Logf("Request for '%s' should have errored - but did not", testCase.input)
				t.Fail()
			}
			continue
		}
		if !shouldError {
			if err != nil {
				t.Logf("Request for '%s' should have not errored - but did: %v", testCase.input, err)
				t.Fail()
			}
			continue
		}

		if len(targets) != 1 {
			t.Logf("%v with %v input => expected one target - got %d", testCase.name, testCase.input, len(targets))
			t.Fail()
			continue
		}
		actualTarget := targets[0]
		expectedTargetExporter := testCase.expect.(*testExporter)

		if actualTarget.exporter != expectedTargetExporter {
			t.Logf("%v with %v input => expected %s target - got %s", testCase.name, testCase.input, testCase.expect.(*testExporter).Name(), actualTarget.exporter.Name())
			t.Fail()
			continue
		}
	}
}
