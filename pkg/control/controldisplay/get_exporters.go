package controldisplay

import (
	"github.com/turbot/steampipe/pkg/export"
)

// GetExporters returns 2 exporter maps, one keyed by name and one keyed by extension
// this is needed because there is some complex logic to resolver control formatter which we do not want to replicate
// so to avoid this we just build the maps and set these on the Export manager directly
func GetExporters() (exportersByName, exportersByExtension map[string]export.Exporter, err error) {
	formatResolver, err := NewFormatResolver()
	if err != nil {
		return nil, nil, err
	}
	exportersByName, exportersByExtension = formatResolver.controlExporters()
	return
}
