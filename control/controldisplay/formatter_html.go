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
	t := template.New("test")
	const temp1 = `<p>Title: {{.Root.Title}}</p>`

	a, err := t.Parse(temp1)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBufferString("")
	res := a.Execute(b, tree)
	if res != nil {
		return nil, res
	}
	output := strings.NewReader(b.String())
	return output, nil
}
