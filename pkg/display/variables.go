package display

import (
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func ShowVarsListJson(vars []*modconfig.Variable) {
	jsonOutput, err := json.MarshalIndent(vars, "", "  ")
	error_helpers.FailOnErrorWithMessage(err, "failed to marshal variables to JSON")

	fmt.Println(string(jsonOutput))
}

func ShowVarsListTable(vars []*modconfig.Variable) {
	headers := []string{"mod_name", "name", "description", "value", "value_default", "type"}
	var rows = make([][]string, len(vars))
	for i, v := range vars {
		rows[i] = []string{v.ModName, v.ShortName, v.Description, fmt.Sprintf("%v", v.ValueGo), fmt.Sprintf("%v", v.DefaultGo), v.TypeString}
	}
	ShowWrappedTable(headers, rows, &ShowWrappedTableOptions{AutoMerge: false})
}
