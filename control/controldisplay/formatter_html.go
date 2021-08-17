package controldisplay

import (
	"context"
	"embed"
	"html/template"
	"io"

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
	reader, writer := io.Pipe()
	go func() {
		res := t.Execute(writer, tree)
		if res != nil {
			writer.CloseWithError(err)
			return
		}
		writer.Close()
	}()
	return reader, nil
}
