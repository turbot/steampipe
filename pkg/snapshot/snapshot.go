package snapshot

import (
	"context"
	"fmt"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/querydisplay"
	"github.com/turbot/pipe-fittings/v2/queryresult"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

const schemaVersion = "20221222"

// PanelData implements SnapshotPanel in the pipe-fittings SteampipeSnapshot struct
// We cannot use the SnapshotPanel interface directly in this package as it references
// powerpipe types that are not available in this package
type PanelData struct {
	Dashboard        string            `json:"dashboard"`
	Name             string            `json:"name"`
	PanelType        string            `json:"panel_type"`
	SourceDefinition string            `json:"source_definition"`
	Status           string            `json:"status,omitempty"`
	Title            string            `json:"title,omitempty"`
	SQL              string            `json:"sql,omitempty"`
	Properties       map[string]string `json:"properties,omitempty"`
	Data             LeafData          `json:"data,omitempty"`
}

type LeafData struct {
	Columns []*queryresult.ColumnDef `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

// IsSnapshotPanel implements SnapshotPanel
func (*PanelData) IsSnapshotPanel() {}

// QueryResultToSnapshot function to generate a snapshot from a query result
func QueryResultToSnapshot[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery, searchPath []string, startTime time.Time) (*steampipeconfig.SteampipeSnapshot, error) {

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

func getPanelDashboard[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) *PanelData {
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

func getPanelTable[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) *PanelData {
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

type snapshotPanelData struct {
	Columns  []*queryresult.ColumnDef `json:"columns"`
	Rows     []map[string]interface{} `json:"rows"`
	Metadata any                      `json:"metadata,omitempty"`
}

func newSnapshotPanelData() *snapshotPanelData {
	return &snapshotPanelData{
		Rows: make([]map[string]interface{}, 0),
	}
}

func getData[T queryresult.TimingContainer](ctx context.Context, result *queryresult.Result[T]) LeafData {
	jsonOutput := newSnapshotPanelData()
	// Ensure columns are being added
	if len(result.Cols) == 0 {
		error_helpers.ShowError(ctx, fmt.Errorf("no columns found in the result"))
	}
	// Add column definitions to the JSON output
	for _, col := range result.Cols {
		c := &pqueryresult.ColumnDef{
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
	return LeafData{
		Columns: jsonOutput.Columns,
		Rows:    jsonOutput.Rows,
	}
}

func getLayout[T queryresult.TimingContainer](result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) *steampipeconfig.SnapshotTreeNode {
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

// SnapshotToQueryResult function to generate a queryresult with streamed rows from a snapshot
func SnapshotToQueryResult[T queryresult.TimingContainer](snap *steampipeconfig.SteampipeSnapshot, startTime time.Time) (*queryresult.Result[T], error) {
	// the table of a snapshot query has a fixed name
	tablePanel, ok := snap.Panels[pconstants.SnapshotQueryTableName]
	if !ok {
		return nil, sperr.New("dashboard does not contain table result for query")
	}
	chartRun := tablePanel.(*PanelData)
	if !ok {
		return nil, sperr.New("failed to read query result from snapshot")
	}

	var tim T
	res := queryresult.NewResult[T](chartRun.Data.Columns, tim)

	// Create a done channel to allow the goroutine to be cancelled
	done := make(chan struct{})

	// start a goroutine to stream the results as rows
	go func() {
		defer res.Close()
		for _, d := range chartRun.Data.Rows {
			// we need to allocate a new slice everytime, since this gets read
			// asynchronously on the other end and we need to make sure that we don't overwrite
			// data already sent
			rowVals := make([]interface{}, len(chartRun.Data.Columns))
			for i, c := range chartRun.Data.Columns {
				rowVals[i] = d[c.Name]
			}

			// Use select with timeout to prevent goroutine leak when consumer stops reading
			select {
			case res.RowChan <- &queryresult.RowResult{Data: rowVals}:
				// Row sent successfully
			case <-done:
				// Cancelled, stop sending rows
				return
			case <-time.After(30 * time.Second):
				// Timeout after 30s - consumer likely stopped reading, exit to prevent leak
				return
			}
		}
	}()

	// Note: The done channel is intentionally not closed anywhere because we don't have
	// a way to detect when the consumer abandons the result. The timeout in the select
	// statement handles the goroutine leak case.

	// res.Timing = &queryresult.TimingMetadata{
	// 	Duration: time.Since(startTime),
	// }
	return res, nil
}
