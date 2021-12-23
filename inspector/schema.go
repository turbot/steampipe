package inspector

import (
	"context"
	"errors"
	"sort"

	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
)

var ErrConnectionNotFound = errors.New("")

func DescribeConnection(ctx context.Context, connectionName string, schemaMetadata schema.Metadata, connectionMap steampipeconfig.ConnectionDataMap) error {
	header := []string{"table", "description"}
	rows := [][]string{}

	schema, found := schemaMetadata.Schemas[connectionName]

	if !found {
		return ErrConnectionNotFound
	}

	for _, tableSchema := range schema {
		rows = append(rows, []string{tableSchema.Name, tableSchema.Description})
	}

	// sort by table name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	display.ShowWrappedTable(header, rows, false)

	return nil
}
