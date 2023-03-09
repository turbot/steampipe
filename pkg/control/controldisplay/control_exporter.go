package controldisplay

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/export"
)

var 	contextKeyFormatterPurpose = contexthelpers.ContextKey("formatter_purpose")
const formatterPurposeExport = "export"

type ControlExporter struct {
	formatter Formatter
}

func NewControlExporter(formatter Formatter) *ControlExporter {
	return &ControlExporter{formatter}
}

func (e *ControlExporter) Export(ctx context.Context, input export.ExportSourceData, destPath string) error {

	// tell the formatter it is being used for export
	// this is a tactical mechanism used to ensure that exported snapshots are unindented
	// whereas display snapshots are indented
	exportCtx := context.WithValue(ctx, contextKeyFormatterPurpose, formatterPurposeExport)

	// input must be control execution tree
	tree, ok := input.(*controlexecute.ExecutionTree)
	if !ok {
		return fmt.Errorf("ControlExporter input must be *controlexecute.ExecutionTree")
	}
	res, err := e.formatter.Format(exportCtx, tree)
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
