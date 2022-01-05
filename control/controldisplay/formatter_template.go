package controldisplay

import (
	"context"
	"io"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

type TemplateFormatter struct {
	outputExtension string
	template        *template.Template
}

func (tf TemplateFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	go func() {
		if err := tf.template.Execute(writer, tree); err != nil {
			writer.CloseWithError(err)
		} else {
			writer.Close()
		}
	}()
	return reader, nil
}

func (tf TemplateFormatter) FileExtension() string {
	return tf.outputExtension
}
