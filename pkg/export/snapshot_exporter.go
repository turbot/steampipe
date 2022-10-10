package export

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"strings"
)

type SnapshotExporter struct {
}

func (e *SnapshotExporter) Export(_ context.Context, input ExportSourceData, filePath string) error {
	snapshot, ok := input.(*dashboardtypes.SteampipeSnapshot)

	if !ok {
		return fmt.Errorf("SnapshotExporter inp-ut must be *dashboardtypes.SteampipeSnapshot")
	}
	snapshotStr, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	res := strings.NewReader(fmt.Sprintf("%s\n", string(snapshotStr)))

	return Write(filePath, res)
}

func (e *SnapshotExporter) FileExtension() string {
	return constants.SnapshotExtension
}

func (e SnapshotExporter) Name() string {
	return constants.OutputFormatSnapshot
}
