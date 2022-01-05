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

func (j MarkdownFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.index.tmpl.md").
		Funcs(formatterTemplateFuncMap).
		ParseFS(mdTemplateFS, "markdown_template/*")
	if err != nil {
		return nil, err
	}
	return TemplateFormatter{template: t}.Format(ctx, tree)
}

func (j MarkdownFormatter) FileExtension() string {
	return "md"
}
