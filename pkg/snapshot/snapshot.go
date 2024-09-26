package snapshot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/querydisplay"
	"github.com/turbot/pipe-fittings/queryresult"
	pqueryresult "github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/pipe-fittings/utils"
)

const schemaVersion = "20221222"

type PanelData struct {
	Dashboard        string                 `json:"dashboard"`
	Name             string                 `json:"name"`
	PanelType        string                 `json:"panel_type"`
	SourceDefinition string                 `json:"source_definition"`
	Status           string                 `json:"status,omitempty"`
	Title            string                 `json:"title,omitempty"`
	SQL              string                 `json:"sql,omitempty"`
	Properties       map[string]string      `json:"properties,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
}

// IsSnapshotPanel implements SnapshotPanel
func (*PanelData) IsSnapshotPanel() {}

type LayoutData struct {
	Name      string        `json:"name"`
	Children  []LayoutChild `json:"children"` // Slice of LayoutChild structs
	PanelType string        `json:"panel_type"`
}

type LayoutChild struct {
	Name      string `json:"name"`
	PanelType string `json:"panel_type"`
}

// QueryResultToSnapshot function to generate a snapshot from a query result
func QueryResultToSnapshot[T any](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery, searchPath []string, startTime time.Time) (*steampipeconfig.SteampipeSnapshot, error) {

	endTime := time.Now()
	hash, err := utils.Base36Hash(resolvedQuery.RawSQL, 8)
	if err != nil {
		return nil, err
	}
	dashboardName := fmt.Sprintf("custom.dashboard.sql_%s", hash)
	// Build the snapshot data (use the new getData function to retrieve data)
	snapshotData := &steampipeconfig.SteampipeSnapshot{
		SchemaVersion: schemaVersion,
		Panels: map[string]steampipeconfig.SnapshotPanel{
			dashboardName:          getPanelDashboard[T](ctx, result, resolvedQuery),
			"custom.table.results": getPanelTable[T](ctx, result, resolvedQuery),
		},
		Inputs:     map[string]interface{}{},
		Variables:  map[string]string{},
		SearchPath: searchPath,
		StartTime:  startTime,
		EndTime:    endTime,
		Layout:     getLayout[T](result, resolvedQuery),
	}
	// Return the snapshot data
	return snapshotData, nil
}

func getPanelDashboard[T any](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) *PanelData {
	hash, err := utils.Base36Hash(resolvedQuery.RawSQL, 8)
	if err != nil {
		return &PanelData{}
	}
	dashboardName := fmt.Sprintf("custom.dashboard.sql_%s", hash)
	// Build panel data with proper fields
	return &PanelData{
		Dashboard:        dashboardName,
		Name:             dashboardName,
		PanelType:        "dashboard",
		SourceDefinition: "",
		Status:           "complete",
		Title:            fmt.Sprintf("Custom query [%s]", hash),
	}
}

func getPanelTable[T any](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) *PanelData {
	hash, err := utils.Base36Hash(resolvedQuery.RawSQL, 8)
	if err != nil {
		return &PanelData{}
	}
	dashboardName := fmt.Sprintf("custom.dashboard.sql_%s", hash)
	// Build panel data with proper fields
	return &PanelData{
		Dashboard:        dashboardName,
		Name:             "custom.table.results",
		PanelType:        "table",
		SourceDefinition: "",
		Status:           "complete",
		SQL:              resolvedQuery.RawSQL,
		Properties: map[string]string{
			"name": "results",
		},
		Data: getData(ctx, result),
	}
}

func getData[T any](ctx context.Context, result *queryresult.Result[T]) map[string]interface{} {
	jsonOutput := querydisplay.NewJSONOutput[T]()
	// Ensure columns are being added
	if len(result.Cols) == 0 {
		error_helpers.ShowError(ctx, fmt.Errorf("no columns found in the result"))
	}
	// Add column definitions to the JSON output
	for _, col := range result.Cols {
		c := pqueryresult.ColumnDef{
			Name:         col.Name,
			OriginalName: col.OriginalName,
			DataType:     strings.ToUpper(col.DataType),
		}
		jsonOutput.Columns = append(jsonOutput.Columns, c)
	}
	// Define function to add each row to the JSON output
	rowFunc := func(row []interface{}, result *queryresult.Result[T]) {
		record := map[string]interface{}{}
		for idx, col := range result.Cols {
			value, _ := querydisplay.ParseJSONOutputColumnValue(row[idx], col)
			record[col.Name] = value
		}
		jsonOutput.Rows = append(jsonOutput.Rows, record)
	}
	// Call iterateResults and ensure rows are processed
	_, err := querydisplay.IterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}
	// Return the full data (including columns and rows)
	return map[string]interface{}{
		"columns": jsonOutput.Columns,
		"rows":    jsonOutput.Rows,
	}
}

func getLayout[T any](result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) *steampipeconfig.SnapshotTreeNode {
	hash, err := utils.Base36Hash(resolvedQuery.RawSQL, 8)
	if err != nil {
		return nil
	}
	dashboardName := fmt.Sprintf("custom.dashboard.sql_%s", hash)
	// Define layout structure
	return &steampipeconfig.SnapshotTreeNode{
		Name: dashboardName,
		Children: []*steampipeconfig.SnapshotTreeNode{
			{
				Name:     "custom.table.results",
				NodeType: "table",
			},
		},
		NodeType: "dashboard",
	}
}
