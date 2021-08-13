package controldisplay

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"strings"

	"github.com/turbot/steampipe/control/controlexecute"
)

type HTMLFormatter struct{}

func (j *HTMLFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	const temp = `<p>Title: {{.Root.Title}}</p>`
	t, err := template.New("test").Parse(temp)
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
