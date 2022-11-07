package export_test

import (
	"context"
	"os"
	"testing"

	"github.com/turbot/steampipe/pkg/control/controldisplay"
	"github.com/turbot/steampipe/pkg/export"
	"github.com/turbot/steampipe/pkg/filepaths"
)

type exporterTestCase struct {
	name        string
	input       []string
	shouldError bool
}

var exporterTestCases = []exporterTestCase{
	{
		name:        "Bad Format",
		input:       []string{"bad-format"},
		shouldError: true,
	},
	{
		name:        "CSV",
		input:       []string{"csv", "file.csv"},
		shouldError: false,
	},
}

func TestDoExport(t *testing.T) {
	deadline, ok := t.Deadline()
	ctx := context.Background()
	if ok {
		newCtx, cancel := context.WithDeadline(ctx, deadline)
		ctx = newCtx
		defer cancel()
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	filepaths.SteampipeDir = tmpDir

	// change the working directory, so that if files get written, they get written
	// to the temp directory which gets removed at the end
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	err = controldisplay.EnsureTemplates()
	if err != nil {
		t.Fatal(err)
	}

	m := export.NewManager()
	ctrlExporters, err := controldisplay.GetExporters()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ctrlExporters {
		m.Register(e)
	}

	for _, testCase := range exporterTestCases {
		err = m.DoExport(ctx, "unimportant", nil, testCase.input)
		if testCase.shouldError && err == nil {
			t.Logf("%v with %v input should have errored, but didn't", testCase.name, testCase.input)
			t.Fail()
			continue
		}
		if !testCase.shouldError && err != nil {
			t.Logf("%v with %v input should not have errored, but errored with %v", testCase.name, testCase.input, err)
			t.Fail()
			continue
		}
	}
}
