package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/turbot/steampipe/control/execute"
)

const (
	OutputFormatNone  = "none"
	OutputFormatText  = "text"
	OutputFormatBrief = "brief"
	OutputFormatCSV   = "csv"
	OutputFormatJSON  = "json"
)

type Formatter interface {
	Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error)
}

func GetFormatter(outputFormat string) (Formatter, error) {
	var formatter Formatter

	switch outputFormat {
	case OutputFormatText, OutputFormatBrief:
		formatter = &TextFormatter{}
	case OutputFormatCSV:
		formatter = &CSVFormatter{}
	case OutputFormatJSON:
		formatter = &JSONFormatter{}
	case OutputFormatNone:
		formatter = &NullFormatter{}
	default:
		return nil, fmt.Errorf("invalid output format '%s' - must be one of json,csv,text,brief,none", outputFormat)
	}

	return formatter, nil
}

// NullFormatter is to be used when no output is expected. It always returns a `io.Reader` which
// reads an empty string
type NullFormatter struct{}

func (j *NullFormatter) Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	return strings.NewReader(""), nil
}
