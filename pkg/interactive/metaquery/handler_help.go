package metaquery

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	localconstants "github.com/turbot/steampipe/pkg/constants"
)

// .help
func doHelp(_ context.Context, _ *HandlerInput) error {
	var commonCmds = []string{localconstants.CmdHelp, localconstants.CmdInspect, localconstants.CmdExit}

	commonCmdRows := getMetaQueryHelpRows(commonCmds, false)
	var advanceCmds []string
	for cmd := range metaQueryDefinitions {
		if !helpers.StringSliceContains(commonCmds, cmd) {
			advanceCmds = append(advanceCmds, cmd)
		}
	}
	advanceCmdRows := getMetaQueryHelpRows(advanceCmds, true)
	// print out
	fmt.Printf("Welcome to Steampipe shell.\n\nTo start, simply enter your SQL query at the prompt:\n\n  select * from aws_iam_user\n\nCommon commands:\n\n%s\n\nAdvanced commands:\n\n%s\n\nDocumentation available at %s\n",
		buildTable(commonCmdRows, true),
		buildTable(advanceCmdRows, true),
		constants.Bold("https://steampipe.io/docs"))
	fmt.Println()
	return nil
}

func getMetaQueryHelpRows(cmds []string, arrange bool) [][]string {
	var rows [][]string
	for _, cmd := range cmds {
		metaQuery := metaQueryDefinitions[cmd]
		var argsStr []string
		if len(metaQuery.args) > 2 {
			rows = append(rows, []string{cmd + " " + "[mode]", metaQuery.description})
		} else {
			for _, v := range metaQuery.args {
				argsStr = append(argsStr, v.value)
			}
			rows = append(rows, []string{cmd + " " + strings.Join(argsStr, "|"), metaQuery.description})
		}
	}
	// sort by metacmds name
	if arrange {
		sort.SliceStable(rows, func(i, j int) bool {
			return rows[i][0] < rows[j][0]
		})
	}
	return rows
}
