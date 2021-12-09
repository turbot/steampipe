package controldisplay

import (
	"context"
	"embed"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type NUnit3Formatter struct{}

//go:embed xml_template/*
var xmlTemplateFS embed.FS

func (j *NUnit3Formatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.index.tmpl.xml").
		Funcs(formatterTemplateFuncMap).
		ParseFS(xmlTemplateFS, "xml_template/*")

	if err != nil {
		return nil, err
	}
	reader, writer := io.Pipe()
	go func() {
		if err := t.Execute(writer, tree); err != nil {
			writer.CloseWithError(err)
		} else {
			writer.Close()
		}
	}()
	return reader, nil
}

func (j *NUnit3Formatter) FileExtension() string {
	return "xml"
}
