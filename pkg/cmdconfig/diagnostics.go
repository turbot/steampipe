package cmdconfig

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"os"
	"strings"
)

func DisplayConfig() {
	diagnostics, ok := os.LookupEnv(constants.EnvDiagnostics)
	if !ok {
		// shouldn't happen
		return
	}
	diagnostics = strings.ToLower(diagnostics)
	configFormats := []string{"config", "config_json"}
	if !helpers.StringSliceContains(configFormats, diagnostics) {
		error_helpers.ShowWarning("invalid value for STEAMPIPE_DIAGNOSTICS, expected values: config,config_json")
		return
	}

	var argNames = []string{
		constants.ArgInstallDir,
		constants.ArgModLocation,
		constants.ArgSnapshotLocation,
		constants.ArgWorkspaceProfile,
		constants.ArgWorkspaceDatabase,
		constants.ArgCloudHost,
		constants.ArgCloudToken,
	}
	res := make(map[string]interface{}, len(argNames))

	maxLength := 0
	for _, a := range argNames {
		if l := len(a); l > maxLength {
			maxLength = l
		}

		res[a] = viper.Get(a)
	}

	switch diagnostics {
	case "config":
		var b strings.Builder
		b.WriteString("\n================\nSteampipe Config\n================\n\n")
		fmtStr := `%-` + fmt.Sprintf("%d", maxLength) + `s: %v` + "\n"

		for k, v := range res {
			b.WriteString(fmt.Sprintf(fmtStr, k, v))
		}
		fmt.Println(b.String())
	case "config_json":
		jsonBytes, err := json.MarshalIndent(res, "", " ")
		error_helpers.FailOnError(err)
		fmt.Println(string(jsonBytes))
	}

}
