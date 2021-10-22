package controldisplay

import (
	"context"
	"embed"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type HTMLFormatter struct{}

//go:embed html_template/*
var templateFS embed.FS

func (j *HTMLFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.index.tmpl.html").
		Funcs(formatterTemplateFuncMap).
		ParseFS(templateFS, "html_template/*")

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

func (j *HTMLFormatter) FileExtension() string {
	return "html"
}
