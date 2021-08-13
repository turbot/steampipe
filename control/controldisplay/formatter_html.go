package controldisplay

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io"
	"strings"

	"github.com/turbot/steampipe/control/controlexecute"
)

type HTMLFormatter struct{}

//go:embed html_template/*
var templateFS embed.FS

func (j *HTMLFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.ParseFS(templateFS, "html_template/*")
	if err != nil {
		return nil, err
	}
	b := bytes.NewBufferString("")
	res := t.Execute(b, tree)
	if res != nil {
		return nil, res
	}
	output := strings.NewReader(b.String())
	return output, nil
}
