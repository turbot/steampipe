package controldisplay

import (
	"testing"
)

type Testcase struct {
	input    string // --export
	expected TemplateFormatterFactoryData
}

var testCases map[string]Testcase = map[string]Testcase{}

var testData = map[string]TemplateFormatterFactoryData{
	"html":              {SourceTemplateFolder: "html", TargetFileExtension: "html"},
	"brief":             {SourceTemplateFolder: "brief.html", TargetFileExtension: "html"},
	"nunit3":            {SourceTemplateFolder: "nunit3.xml", TargetFileExtension: "xml"},
	"markdown":          {SourceTemplateFolder: "markdown.md", TargetFileExtension: "md"},
	"txt":               {SourceTemplateFolder: "txt.dat", TargetFileExtension: "dat"},
	"foo":               {SourceTemplateFolder: "foo.xml", TargetFileExtension: "xml"},
	"brief.html":        {SourceTemplateFolder: "brief.html", TargetFileExtension: "html"},
	"nunit3.xml":        {SourceTemplateFolder: "nunit3.xml", TargetFileExtension: "xml"},
	"markdown.md":       {SourceTemplateFolder: "markdown.md", TargetFileExtension: "md"},
	"txt.dat":           {SourceTemplateFolder: "txt.dat", TargetFileExtension: "dat"},
	"custom.txt":        {SourceTemplateFolder: "custom.txt", TargetFileExtension: "txt"},
	"foo.xml":           {SourceTemplateFolder: "foo.xml", TargetFileExtension: "xml"},
	"output.html":       {SourceTemplateFolder: "html", TargetFileExtension: "html"},
	"output.md":         {SourceTemplateFolder: "markdown.md", TargetFileExtension: "md"},
	"output.txt":        {SourceTemplateFolder: "custom.txt", TargetFileExtension: "txt"},
	"output.dat":        {SourceTemplateFolder: "txt.dat", TargetFileExtension: "dat"},
	"output.brief.html": {SourceTemplateFolder: "brief.html", TargetFileExtension: "html"},
	"output.nunit3.xml": {SourceTemplateFolder: "nunit3.xml", TargetFileExtension: "xml"},
	"output.foo.xml":    {SourceTemplateFolder: "foo.xml", TargetFileExtension: "xml"},
}

func TestTemplateExport(t *testing.T) {
	for i, d := range testData {
		fff, err := GetTemplateFormatter(i, "all")
		if err != nil {
			t.Error(err)
		}

		if d.SourceTemplateFolder != fff.SourceTemplateFolder || d.TargetFileExtension != fff.TargetFileExtension {
			t.Logf(`"expected:%v" is not equal to "output:%v"`, d, fff)
		}
	}
}
