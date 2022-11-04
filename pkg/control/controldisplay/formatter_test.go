package controldisplay

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

var rootBenchmark = modconfig.Benchmark{}
var childBenchmark1 = modconfig.Benchmark{}
var childBenchmark2 = modconfig.Benchmark{}

var desc = "Dummy control for unit testing"
var title = "DummyControl"
var c11 = modconfig.Control{
	Title:       &title,
	Description: &desc,
}
var c12 = modconfig.Control{
	Title:       &title,
	Description: &desc,
}
var c21 = modconfig.Control{
	Title:       &title,
	Description: &desc,
}
var c22 = modconfig.Control{
	Title:       &title,
	Description: &desc,
}

var tree = &controlexecute.ExecutionTree{
	Root: &controlexecute.ResultGroup{
		GroupId:     "DummyTest",
		Parent:      nil,
		Title:       "Test Root Group",
		Description: "Description for test root group",
		Summary: &controlexecute.GroupSummary{
			Status: controlstatus.StatusSummary{
				Alarm: 2,
			},
		},
		GroupItem: &rootBenchmark,
		Groups: []*controlexecute.ResultGroup{
			{
				GroupItem: &childBenchmark1,
				ControlRuns: []*controlexecute.ControlRun{
					{
						Control: &c11,
						Rows: []*controlexecute.ResultRow{
							{
								Status:     constants.ControlAlarm,
								Reason:     "is pretty insecure",
								Resource:   "some other resource",
								Dimensions: []controlexecute.Dimension{},
								Run:        &controlexecute.ControlRun{Control: &c11},
							},
						},
					},
					{
						Control: &c12,
						Rows: []*controlexecute.ResultRow{
							{
								Status:     constants.ControlAlarm,
								Reason:     "is pretty insecure",
								Resource:   "some other resource",
								Dimensions: []controlexecute.Dimension{},
								Run:        &controlexecute.ControlRun{Control: &c12},
							},
						},
					},
				},
			},
			{
				GroupItem: &childBenchmark2,
				ControlRuns: []*controlexecute.ControlRun{
					{
						Control: &c21,
						Rows: []*controlexecute.ResultRow{
							{
								Status:     constants.ControlAlarm,
								Reason:     "is pretty insecure",
								Resource:   "some other resource",
								Dimensions: []controlexecute.Dimension{},
								Run:        &controlexecute.ControlRun{Control: &c21},
							},
						},
					},
					{
						Control: &c22,
						Rows: []*controlexecute.ResultRow{
							{
								Status:     constants.ControlAlarm,
								Reason:     "is pretty insecure",
								Resource:   "some other resource",
								Dimensions: []controlexecute.Dimension{},
								Run:        &controlexecute.ControlRun{Control: &c22},
							},
						},
					},
				},
			},
		},
	},
}

type exporterTest struct {
	shouldError bool
	alias       string
	extension   string
	name        string
}

// testFormatter is an implementation of the Formatter interface
// values in this implementation correspond to the ones we expect in the result
type testFormatter struct {
	name      string
	alias     string
	extension string
}

func (b *testFormatter) FileExtension() string { return b.extension }
func (b *testFormatter) Name() string          { return b.name }
func (b *testFormatter) Alias() string         { return b.alias }
func (b *testFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	return nil, nil
}

type testCase struct {
	input    string
	expected interface{}
}

var exporterTestCases = []testCase{
	{
		input:    "bad-format",
		expected: "ERROR",
	},
	{
		input: "snapshot",
		expected: testFormatter{
			alias:     "sps",
			extension: constants.SnapshotExtension,
			name:      constants.OutputFormatSnapshot,
		},
	},
	{
		input: "csv",
		expected: testFormatter{
			alias:     "",
			extension: ".csv",
			name:      "csv",
		},
	},
	{
		input: "json",
		expected: testFormatter{
			alias:     "",
			extension: ".json",
			name:      "json",
		},
	},
}

func TestFormatResolver(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	filepaths.SteampipeDir = tmpDir
	if err := EnsureTemplates(); err != nil {
		t.Fatal(err)
	}
	resolver, err := NewFormatResolver()
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range exporterTestCases {
		f, ferr := resolver.GetFormatter(testCase.input)
		shouldError := testCase.expected == "ERROR"

		if shouldError {
			if ferr == nil {
				t.Logf("Request for '%s' should have errored - but did not", testCase.input)
				t.Fail()
			}
			continue
		}

		expectedFormatter := testCase.expected.(testFormatter)

		if f.Alias() != expectedFormatter.Alias() {
			t.Logf("Alias mismatch for '%s'. Expected '%s', but got '%s'", testCase.input, expectedFormatter.Alias(), f.Alias())
			t.Fail()
			continue
		}
		if f.FileExtension() != expectedFormatter.FileExtension() {
			t.Logf("Extension mismatch for '%s'. Expected '%s', but got '%s'", testCase.input, expectedFormatter.FileExtension(), f.FileExtension())
			t.Fail()
			continue
		}
		if f.Name() != expectedFormatter.Name() {
			t.Logf("Name mismatch for '%s'. Expected '%s', but got '%s'", testCase.input, expectedFormatter.Name(), f.Name())
			t.Fail()
			continue
		}
	}
}
