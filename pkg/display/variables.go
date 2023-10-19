package display

import (
	"encoding/json"
	"fmt"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

type variableInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"value_default"`
	Value       any    `json:"value"`
	ModName     string `json:"mod_name"`
}

func ShowVarsListJson(vars []*modconfig.Variable) {
	var jsonStructs []variableInfo
	for _, v := range vars {
		jv := variableInfo{
			Name:        v.ShortName,
			Type:        v.TypeString,
			Description: v.GetDescription(),
			Default:     v.DefaultGo,
			Value:       v.ValueGo,
			ModName:     v.ModName,
		}
		jsonStructs = append(jsonStructs, jv)
	}
	jsonOutput, err := json.MarshalIndent(jsonStructs, "", "  ")
	error_helpers.FailOnErrorWithMessage(err, "failed to marshal variables to JSON")

	fmt.Println(string(jsonOutput))
}

func ShowVarsListTable(vars []*modconfig.Variable) {
	headers := []string{"mod_name", "name", "description", "value", "value_default", "type"}
	var rows = make([][]string, len(vars))
	for i, v := range vars {
		rows[i] = []string{v.ModName, v.ShortName, v.GetDescription(), fmt.Sprintf("%v", v.ValueGo), fmt.Sprintf("%v", v.DefaultGo), v.TypeString}
	}
	ShowWrappedTable(headers, rows, &ShowWrappedTableOptions{AutoMerge: false})
}
