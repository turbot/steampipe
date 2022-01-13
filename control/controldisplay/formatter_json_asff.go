package controldisplay

import (
	"context"
	"embed"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type AsffJsonFormatter struct{}

//go:embed json_asff_template/*
var asffJsonTemplateFS embed.FS

func (j AsffJsonFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.asff.tmpl.json").
		Funcs(formatterTemplateFuncMap).
		ParseFS(asffJsonTemplateFS, "json_asff_template/*")
	if err != nil {
		return nil, err
	}

	return TemplateFormatter{template: t}.Format(ctx, tree)
}

func (j AsffJsonFormatter) FileExtension() string {
	return "json"
}
