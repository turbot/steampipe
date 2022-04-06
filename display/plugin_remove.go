package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type PluginRemoveReport struct {
	FullPluginName string
	ShortName      string
	Connections    []modconfig.Connection
}

type PluginRemoveReports []PluginRemoveReport

func (prr PluginRemoveReports) Print() {
	length := len(prr)
	staleConnections := []modconfig.Connection{}
	if length > 0 {
		fmt.Printf("\nUninstalled %s:\n", utils.Pluralize("plugin", length))
		for _, report := range prr {
			fmt.Printf("* %s\n", report.ShortName)
			staleConnections = append(staleConnections, report.Connections...)
			sort.SliceStable(report.Connections, func(i, j int) bool {
				return report.Connections[j].DeclRange.Start.Line < report.Connections[i].DeclRange.Start.Line
			})
		}
		fmt.Println()
		staleLength := len(staleConnections)
		uniqueFiles := map[string]bool{}
		// get the unique files
		if staleLength > 0 {
			for _, report := range prr {
				for _, conn := range report.Connections {
					uniqueFiles[conn.DeclRange.Filename] = true
				}
			}

			str := append([]string{}, fmt.Sprintf(
				"Please remove %s %s to continue using steampipe:",
				utils.Pluralize("this", len(uniqueFiles)),
				utils.Pluralize("connection", len(uniqueFiles)),
			))

			str = append(str, "")

			for file := range uniqueFiles {
				str = append(str, fmt.Sprintf("  * %s contains:", constants.Bold(file)))
				for _, report := range prr {
					for _, conn := range report.Connections {
						if conn.DeclRange.Filename == file {
							str = append(str, fmt.Sprintf("         '%s' in line %2d", conn.Name, conn.DeclRange.Start.Line))
						}
					}
				}
				str = append(str, "")
			}

			fmt.Println(strings.Join(str, "\n"))
		}
	}
}
