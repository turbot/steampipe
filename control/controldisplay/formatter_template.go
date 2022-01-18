package controldisplay

import (
	"context"
	"errors"
	"io"
	"os"
	"text/template"

	"github.com/turbot/steampipe/control/controlexecute"
)

var ErrAmbiguousTemplate = errors.New("ambiguous templates found")
var ErrTemplateNotFound = errors.New("template not found")

// TemplateFormatter implements the 'Formatter' interface and exposes a generic template based output mechanism
// for 'check' execution trees
type TemplateFormatter struct {
	template     *template.Template
	exportFormat ExportTemplate
}

func (tf TemplateFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	go func() {
		if err := tf.template.ExecuteTemplate(writer, "outlet", tree); err != nil {
			writer.CloseWithError(err)
		} else {
			writer.Close()
		}
	}()
	return reader, nil
}

func (tf TemplateFormatter) FileExtension() string {
	// if the extension is the same as the format name, return just the extension
	if tf.exportFormat.DefaultTemplateForExtension {
		return tf.exportFormat.OutputExtension
	} else {
		// otherwise return the fullname
		return tf.exportFormat.FormatFullName
	}
}

func NewTemplateFormatter(input ExportTemplate) (*TemplateFormatter, error) {
	t := template.Must(template.New("outlet").
		Funcs(formatterTemplateFuncMap).
		ParseFS(os.DirFS(input.TemplatePath), "*"))

	return &TemplateFormatter{exportFormat: input, template: t}, nil
}
