package controldisplay

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/turbot/steampipe/filepaths"
)

type Testcase struct {
	input    string // --export <val>
	expected interface{}
}

func setup() {
	filepaths.SteampipeDir = "~/.steampipe"
	source, err := filepath.Abs("templates")
	if err != nil {
		log.Fatal(err)
	}

	dest, err := filepath.Abs("~/.steampipe/check/templates")
	if err != nil {
		log.Fatal(err)
	}

	err = copy.Copy(source, dest)
	if err != nil {
		log.Fatal(err)
	}
}

func teardown() {
	os.RemoveAll("~/.steampipe/check/templates")
}

func TestExportFormat(t *testing.T) {
	setup()
	for name, test := range exportFormatTestCases {
		fff, _, err := ResolveExportTemplate(test.input, true)
		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
			}
			continue
		}
		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}
		expectedFormat := test.expected.(ExportTemplate)
		if !FormatEqual(fff, &expectedFormat) {
			t.Errorf("Test: '%s'' FAILED : expected:\n%s\n\ngot:\n%s", name, expectedFormat, fff)
		}
	}
	teardown()
}

func FormatEqual(l, r *ExportTemplate) bool {
	return (l.FormatFullName == r.FormatFullName)
}

var exportFormatTestCases map[string]Testcase = map[string]Testcase{
	"html": {
		input: "html",
		expected: ExportTemplate{
			FormatFullName:  "html.html",
			OutputExtension: ".html",
		},
	},
	"nunit3": {
		input: "nunit3",
		expected: ExportTemplate{
			FormatFullName:  "nunit3.xml",
			OutputExtension: ".xml",
		},
	},
	"markdown": {
		input: "md",
		expected: ExportTemplate{
			FormatFullName:  "md.md",
			OutputExtension: ".md",
		},
	},
	"brief.html": {
		input: "brief.html",
		expected: ExportTemplate{
			FormatFullName:  "html.html",
			OutputExtension: ".html",
		},
	},
	"nunit3.xml": {
		input: "nunit3.xml",
		expected: ExportTemplate{
			FormatFullName:  "nunit3.xml",
			OutputExtension: ".xml",
		},
	},
	"markdown.md": {
		input: "markdown.md",
		expected: ExportTemplate{
			FormatFullName:  "md.md",
			OutputExtension: ".md",
		},
	},
	// "txt.dat": {
	// 	input: "txt.dat",
	// 	expected: ExportTemplate{
	// 		FormatFullName:  "txt.dat",
	// 		OutputExtension: ".dat",
	// 	},
	// },
	// "custom.txt": {
	// 	input: "custom.txt",
	// 	expected: ExportTemplate{
	// 		FormatFullName:  "custom.txt",
	// 		OutputExtension: ".txt",
	// 	},
	// },
	"foo.xml": {
		input: "foo.xml",
		expected: ExportTemplate{
			FormatFullName:  "nunit3.xml",
			OutputExtension: ".xml",
		},
	},
	"output.html": {
		input: "output.html",
		expected: ExportTemplate{
			FormatFullName:  "html.html",
			OutputExtension: ".html",
		},
	},
	"output.md": {
		input: "output.md",
		expected: ExportTemplate{
			FormatFullName:  "md.md",
			OutputExtension: ".md",
		},
	},
	// "output.txt": {
	// 	input: "output.txt",
	// 	expected: ExportTemplate{
	// 		FormatFullName:  "custom.txt",
	// 		OutputExtension: ".txt",
	// 	},
	// },
	// "output.dat": {
	// 	input: "output.dat",
	// 	expected: ExportTemplate{
	// 		FormatFullName:  "txt.dat",
	// 		OutputExtension: ".dat",
	// 	},
	// },
	"output.brief.html": {
		input: "output.brief.html",
		expected: ExportTemplate{
			FormatFullName:  "html.html",
			OutputExtension: ".html",
		},
	},
	"output.nunit3.xml": {
		input: "output.nunit3.xml",
		expected: ExportTemplate{
			FormatFullName:  "nunit3.xml",
			OutputExtension: ".xml",
		},
	},
	"output.foo.xml": {
		input: "output.foo.xml",
		expected: ExportTemplate{
			FormatFullName:  "nunit3.xml",
			OutputExtension: ".xml",
		},
	},
}
