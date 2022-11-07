package controldisplay

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/filepaths"
)

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

var formatterTestCase = []testCase{
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
	{
		input: "asff.json",
		expected: testFormatter{
			alias:     "asff.json",
			extension: ".asff.json",
			name:      "asff",
		},
	},
	{
		input: "nunit3.xml",
		expected: testFormatter{
			alias:     "nunit3.xml",
			extension: ".nunit3.xml",
			name:      "nunit3",
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
	for _, testCase := range formatterTestCase {
		f, err := resolver.GetFormatter(testCase.input)
		shouldError := testCase.expected == "ERROR"

		if shouldError {
			if err == nil {
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
