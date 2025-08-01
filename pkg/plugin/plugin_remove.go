package plugin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/pipe-fittings/v2/utils"
)

type PluginRemoveReport struct {
	Image       *ociinstaller.ImageRef
	ShortName   string
	Connections []PluginConnection
}

type PluginRemoveReports []PluginRemoveReport

func (r PluginRemoveReports) Print() {
	length := len(r)
	var staleConnections []PluginConnection
	if length > 0 {
		fmt.Printf("\nUninstalled %s:\n", utils.Pluralize("plugin", length)) //nolint:forbidigo // acceptable
		for _, report := range r {
			org, name, _ := report.Image.GetOrgNameAndStream()
			fmt.Printf("* %s/%s\n", org, name) //nolint:forbidigo // acceptable
			staleConnections = append(staleConnections, report.Connections...)

			// sort the connections by line number while we are at it!
			sort.SliceStable(report.Connections, func(i, j int) bool {
				left := report.Connections[i]
				right := report.Connections[j]
				return left.GetDeclRange().Start.Line < right.GetDeclRange().Start.Line
			})
		}
		fmt.Println() //nolint:forbidigo // acceptable
		staleLength := len(staleConnections)
		uniqueFiles := map[string]bool{}
		// get the unique files
		if staleLength > 0 {
			for _, report := range r {
				for _, conn := range report.Connections {
					uniqueFiles[conn.GetDeclRange().Filename] = true
				}
			}

			str := append([]string{}, fmt.Sprintf(
				"The following %s %s no longer needed since %s %s been uninstalled and can be safely removed:",
				utils.Pluralize("connection", len(uniqueFiles)),
				utils.Pluralize("is", len(uniqueFiles)),
				utils.Pluralize("the associated plugin", len(uniqueFiles)),
				utils.Pluralize("has", len(uniqueFiles)),
			))

			str = append(str, "")

			for file := range uniqueFiles {
				str = append(str, fmt.Sprintf("  * %s", constants.Bold(file)))
				for _, report := range r {
					for _, conn := range report.Connections {
						if conn.GetDeclRange().Filename == file {
							str = append(str, fmt.Sprintf("         '%s' (line %2d)", conn.GetName(), conn.GetDeclRange().Start.Line))
						}
					}
				}
				str = append(str, "")
			}

			fmt.Println(strings.Join(str, "\n")) //nolint:forbidigo // acceptable
		}
	}
}
