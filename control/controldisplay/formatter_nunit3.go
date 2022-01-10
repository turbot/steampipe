package controldisplay

import (
	"context"
	"embed"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type Nunit3Formatter struct{}

//go:embed nunit3_template/*
var nunit3TemplateFS embed.FS

func (j Nunit3Formatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.nunit3.tmpl.xml").
		Funcs(formatterTemplateFuncMap).
		ParseFS(nunit3TemplateFS, "nunit3_template/*")
	if err != nil {
		return nil, err
	}

	return TemplateFormatter{template: t}.Format(ctx, tree)
}

func (j Nunit3Formatter) FileExtension() string {
	return "xml"
}
