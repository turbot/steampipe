package controldisplay

import "github.com/turbot/steampipe/pkg/export"

func GetExporters() ([]export.Exporter, error) {
	formatResolver, err := NewFormatResolver()
	if err != nil {
		return nil, err
	}
	exporters := formatResolver.controlExporters()
	return exporters, nil
}
