package inspector

import (
	"context"
	"fmt"
	"sort"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
)

func ListConnections(ctx context.Context, schemaMetadata schema.Metadata, connectionMap steampipeconfig.ConnectionDataMap) error {
	header := []string{"connection", "plugin"}
	rows := [][]string{}

	for schema := range schemaMetadata.Schemas {
		plugin, found := connectionMap[schema]
		if found {
			rows = append(rows, []string{schema, plugin.Plugin})
		} else {
			rows = append(rows, []string{schema, ""})
		}
	}

	// sort by connection name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	display.ShowWrappedTable(header, rows, false)

	fmt.Printf(`
To get information about the tables in a connection, run %s
To get information about the columns in a table, run %s

`, constants.Bold(".inspect {connection}"), constants.Bold(".inspect {connection}.{table}"))

	return nil
}
