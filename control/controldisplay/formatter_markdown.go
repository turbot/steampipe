package controldisplay

import (
	"context"
	"embed"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type MarkdownFormatter struct{}

//go:embed markdown_template/*
var mdTemplateFS embed.FS

func (j *MarkdownFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.index.tmpl.md").
		Funcs(formatterTemplateFuncMap).
		ParseFS(mdTemplateFS, "markdown_template/*")
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

func (j *MarkdownFormatter) FileExtension() string {
	return "md"
}
