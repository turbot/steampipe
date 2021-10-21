package controldisplay

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/version"
)

type MarkdownFormatter struct{}

//go:embed markdown_template/*
var mdTemplateFS embed.FS

func (j *MarkdownFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	t, err := template.
		New("001.index.tmpl.md").
		Funcs(template.FuncMap{
			"steampipeversion": func() string { return version.String() },
			"workingdir":       func() string { wd, _ := os.Getwd(); return wd },
			"asstr":            func(i reflect.Value) string { return fmt.Sprintf("%v", i) },
			"statusicon": func(status string) string {
				switch strings.ToLower(status) {
				case "ok":
					return "✅"
				case "skip":
					return "⇨"
				case "info":
					return "ℹ"
				case "alarm":
					return "❌"
				case "error":
					return "❗"
				}
				return ""
			},
		}).
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
