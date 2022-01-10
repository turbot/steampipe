package controldisplay

import (
	"context"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type TemplateFormatter struct {
	outputExtension string
	template        *template.Template
}

func (tf TemplateFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	go func() {
		if err := tf.template.Execute(writer, tree); err != nil {
			writer.CloseWithError(err)
		} else {
			writer.Close()
		}
	}()
	return reader, nil
}

func (tf TemplateFormatter) FileExtension() string {
	return tf.outputExtension
}

// var AvailableFolders = []string{
// 	// this list will be extracted from the filesystem
// 	"html",
// 	"brief.html",
// 	"nunit3.xml",
// 	"markdown.md",
// 	"txt.dat",
// 	"custom.txt",
// 	"foo.xml",
// }

// func GetTemplateFormatter(export string, input string) (*TemplateFormatterFactoryData, error) {
// 	fff := TemplateFormatterFactoryData{}

// 	for _, folder := range AvailableFolders {
// 		// logic
// 	}

// 	return &fff, nil
// }

// type TemplateFormatterFactoryData struct {
// 	SourceTemplateFolder string
// 	TargetFileName       string
// 	TargetFileExtension  string
// }

// given a set of available folder names
