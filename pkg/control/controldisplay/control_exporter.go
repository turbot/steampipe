package controldisplay

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/control/controlexecute"

	"github.com/turbot/steampipe/pkg/export"
)

type ControlExporter struct {
	formatter Formatter
}

func NewControlExporter(formatter Formatter) *ControlExporter {
	return &ControlExporter{formatter}
}

func (e *ControlExporter) Export(ctx context.Context, input export.ExportSourceData, destPath string) error {
	// input must be control execution tree
	tree, ok := input.(*controlexecute.ExecutionTree)
	if !ok {
		return fmt.Errorf("ControlExporter input must be *controlexecute.ExecutionTree")
	}
	res, err := e.formatter.Format(ctx, tree)
	if err != nil {
		return err
	}

	return export.Write(destPath, res)
}

func (e *ControlExporter) FileExtension() string {
	return e.formatter.FileExtension()
}

func (e *ControlExporter) Name() string {
	return e.formatter.Name()
}

func (e *ControlExporter) Alias() string {
	return e.formatter.Alias()
}

func (e *ControlExporter) IsDefaultExporterForExtension() bool {
	return e.formatter.IsDefaultFormatterForExtension()
}
