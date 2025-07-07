package cmdconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

// DisplayConfig prints all config set via WorkspaceProfile or HCL options
func DisplayConfig() {
	diagnostics, ok := os.LookupEnv(constants.EnvConfigDump)
	if !ok {
		// shouldn't happen
		return
	}
	diagnostics = strings.ToLower(diagnostics)
	configFormats := []string{"config", "config_json"}
	if !slices.Contains(configFormats, diagnostics) {
		error_helpers.ShowWarning("invalid value for STEAMPIPE_CONFIG_DUMP, expected values: config,config_json")
		return
	}

	var configArgNames = viper.AllKeys()
	res := make(map[string]interface{}, len(configArgNames))

	maxLength := 0
	for _, a := range configArgNames {
		if l := len(a); l > maxLength {
			maxLength = l
		}
		res[a] = viper.Get(a)
	}

	switch diagnostics {
	case "config":
		// write config lines into array then sort them
		lines := make([]string, len(res))
		idx := 0
		fmtStr := `%-` + fmt.Sprintf("%d", maxLength) + `s: %v` + "\n"
		for k, v := range res {
			lines[idx] = fmt.Sprintf(fmtStr, k, v)
			idx++
		}
		sort.Strings(lines)

		var b strings.Builder
		b.WriteString("\n================\nSteampipe Config\n================\n\n")

		for _, line := range lines {
			b.WriteString(line)
		}
		fmt.Println(b.String())
	case "config_json":
		// iterate once more for the non-serializable values
		for k, v := range res {
			if _, err := json.Marshal(v); err != nil {
				res[k] = fmt.Sprintf("%v", v)
			}
		}
		jsonBytes, err := json.MarshalIndent(res, "", " ")
		error_helpers.FailOnError(err)
		fmt.Println(string(jsonBytes))
	}

}
