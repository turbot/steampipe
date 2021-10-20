package controldisplay

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/control/controlexecute"
)

type FormatterMap map[string]Formatter

func (m FormatterMap) keys() []string {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	return keys
}

const (
	OutputFormatNone     = "none"
	OutputFormatText     = "text"
	OutputFormatBrief    = "brief"
	OutputFormatCSV      = "csv"
	OutputFormatJSON     = "json"
	OutputFormatHTML     = "html"
	OutputFormatMarkdown = "markdown"
)

var outputFormatters FormatterMap = FormatterMap{
	OutputFormatNone:  &NullFormatter{},
	OutputFormatCSV:   &CSVFormatter{},
	OutputFormatJSON:  &JSONFormatter{},
	OutputFormatText:  &TextFormatter{},
	OutputFormatBrief: &TextFormatter{},
}

var exportFormatters FormatterMap = FormatterMap{
	OutputFormatCSV:      &CSVFormatter{},
	OutputFormatJSON:     &JSONFormatter{},
	OutputFormatHTML:     &HTMLFormatter{},
	OutputFormatMarkdown: &MarkdownFormatter{},
}

type CheckExportTarget struct {
	Format string
	File   string
	Error  error
}

func NewCheckExportTarget(format string, file string, err error) CheckExportTarget {
	return CheckExportTarget{
		Format: format,
		File:   file,
		Error:  err,
	}
}

type Formatter interface {
	Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error)
}

func GetExportFormatter(exportFormat string) (Formatter, error) {
	formatter, found := exportFormatters[exportFormat]
	if !found {
		return nil, fmt.Errorf("invalid export format '%s' - must be one of %s", exportFormat, exportFormatters.keys())
	}
	return formatter, nil
}

func GetOutputFormatter(outputFormat string) (Formatter, error) {
	formatter, found := outputFormatters[outputFormat]
	if !found {
		return nil, fmt.Errorf("invalid output format '%s' - must be one of %s", outputFormat, outputFormatters.keys())
	}
	return formatter, nil
}

func InferFormatFromExportFileName(filename string) (string, error) {
	extension := strings.TrimPrefix(filepath.Ext(filename), ".")
	switch extension {
	case "csv":
		return OutputFormatCSV, nil
	case "json":
		return OutputFormatJSON, nil
	case "html":
		return OutputFormatHTML, nil
	case "md", "markdown":
		return OutputFormatMarkdown, nil
	default:
		// return blank, so that it fails when it looks
		// up the formatter when it's to format
		return "", fmt.Errorf("could not infer valid export format from filename '%s'", filename)
	}
}

// NullFormatter is to be used when no output is expected. It always returns a `io.Reader` which
// reads an empty string
type NullFormatter struct{}

func (j *NullFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	return strings.NewReader(""), nil
}
