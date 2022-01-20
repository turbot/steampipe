package controldisplay

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/version"
)

type TemplateRenderConfig struct {
	RenderHeader bool
}
type TemplateRenderConstants struct {
	SteampipeVersion string
}

type TemplateRenderContext struct {
	Constants TemplateRenderConstants
	Config    TemplateRenderConfig
	Data      *controlexecute.ExecutionTree
}

// TemplateFormatter implements the 'Formatter' interface and exposes a generic template based output mechanism
// for 'check' execution trees
type TemplateFormatter struct {
	template     *template.Template
	exportFormat ExportTemplate
}

func (tf TemplateFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	go func() {
		renderContext := TemplateRenderContext{
			Constants: TemplateRenderConstants{
				SteampipeVersion: version.SteampipeVersion.String(),
			},
			Config: TemplateRenderConfig{
				RenderHeader: viper.GetBool(constants.ArgHeader),
			},
			Data: tree,
		}

		// overwrite the "render_context" function to return the current render context
		templateFuncs := templateFuncs()
		templateFuncs["render_context"] = func() TemplateRenderContext { return renderContext }

		t, err := tf.template.Clone()
		if err != nil {
			writer.CloseWithError(err)
			return
		}
		t = t.Funcs(templateFuncs)

		if err := t.ExecuteTemplate(writer, "output", renderContext); err != nil {
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
		return fmt.Sprintf(".%s", tf.exportFormat.FormatFullName)
	}
}

func NewTemplateFormatter(input ExportTemplate) (*TemplateFormatter, error) {
	templateFuncs := templateFuncs()

	// add a stub "render_context" function
	// this will be overwritten before we execute the template
	// if we don't put this here, then templates which use this
	// won't parse and will throw Error: template: ****: function "render_context" not defined
	templateFuncs["render_context"] = func() TemplateRenderContext { return TemplateRenderContext{} }

	t := template.Must(template.New("outlet").
		Funcs(templateFuncs).
		ParseFS(os.DirFS(input.TemplatePath), "*"))

	return &TemplateFormatter{exportFormat: input, template: t}, nil
}
