package controldisplay

import (
	"context"
	"github.com/turbot/steampipe/pkg/utils"
	"io"
	"os"
	"text/template"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/version"
)

// TemplateFormatter implements the 'Formatter' interface and exposes a generic template based output mechanism
// for 'check' execution trees
type TemplateFormatter struct {
	template     *template.Template
	exportFormat *OutputTemplate
}

func NewTemplateFormatter(input *OutputTemplate) (*TemplateFormatter, error) {
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

func (tf TemplateFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	reader, writer := io.Pipe()
	go func() {
		workingDirectory, err := os.Getwd()
		if err != nil {
			writer.CloseWithError(err)
			return
		}
		renderContext := TemplateRenderContext{
			Constants: TemplateRenderConstants{
				SteampipeVersion: version.SteampipeVersion.String(),
				WorkingDir:       workingDirectory,
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

	// tactical - for json, prettify the output
	if tf.shouldPrettify(){
		return utils.PrettifyJsonFromReader(reader)
	}

	return reader, nil
}

func (tf TemplateFormatter) FileExtension() string {
	return tf.exportFormat.FileExtension
}

func (tf TemplateFormatter) Name() string {
	return tf.exportFormat.FormatName
}

func (tf TemplateFormatter) Alias() string {
	if tf.exportFormat.FormatFullName != tf.exportFormat.FormatName {
		return tf.exportFormat.FormatFullName
	}
	return ""
}

func (tf TemplateFormatter) shouldPrettify() bool {
	return tf.Name() == constants.OutputFormatJSON
}
