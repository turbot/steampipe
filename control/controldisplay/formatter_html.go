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

	return TemplateFormatter{template: t}.Format(ctx, tree)
}

func (j *HTMLFormatter) FileExtension() string {
	return "html"
}
