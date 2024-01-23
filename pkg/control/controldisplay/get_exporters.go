package controldisplay

import (
	"context"

	"github.com/turbot/steampipe/pkg/export"
)

// GetExporters returns an array of ControlExporters corresponding to the available output formats
func GetExporters(ctx context.Context) ([]export.Exporter, error) {
	formatResolver, err := NewFormatResolver(ctx)
	if err != nil {
		return nil, err
	}
	exporters := formatResolver.controlExporters()
	return exporters, nil
}
