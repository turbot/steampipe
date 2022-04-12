package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/utils"
)

type PluginInstallReports []*PluginInstallReport

// making the type compatible with sort.Interface so that we can use the sort package utilities
func (ir PluginInstallReports) Len() int           { return len(ir) }
func (ir PluginInstallReports) Swap(i, j int)      { ir[i], ir[j] = ir[j], ir[i] }
func (ir PluginInstallReports) Less(i, j int) bool { return ir[i].Plugin < ir[j].Plugin }

type PluginInstallReport struct {
	Skipped        bool
	Plugin         string
	SkipReason     string
	DocURL         string
	Version        string
	IsUpdateReport bool
}

func (i *PluginInstallReport) skipString() string {
	ref := ociinstaller.NewSteampipeImageRef(i.Plugin)
	_, name, stream := ref.GetOrgNameAndStream()

	return fmt.Sprintf("Plugin:   %s\nReason:   %s", fmt.Sprintf("%s@%s", name, stream), i.SkipReason)
}

func (i *PluginInstallReport) installString() string {
	thisReport := []string{}
	if i.IsUpdateReport {
		thisReport = append(
			thisReport,
			fmt.Sprintf("Updated plugin: %s%s", constants.Bold(i.Plugin), i.Version),
		)
		if len(i.DocURL) > 0 {
			thisReport = append(
				thisReport,
				fmt.Sprintf("Documentation:  %s", i.DocURL),
			)
		}
	} else {
		thisReport = append(
			thisReport,
			fmt.Sprintf("Installed plugin: %s%s", constants.Bold(i.Plugin), i.Version),
		)
		if len(i.DocURL) > 0 {
			thisReport = append(
				thisReport,
				fmt.Sprintf("Documentation:    %s", i.DocURL),
			)
		}
	}

	return strings.Join(thisReport, "\n")
}

func (i *PluginInstallReport) String() string {
	if !i.Skipped {
		return i.installString()
	} else {
		return i.skipString()
	}
}

// PrintInstallReports Prints out the installation reports onto the console
func PrintInstallReports(reports PluginInstallReports, isUpdateReport bool) {
	installedOrUpdated := PluginInstallReports{}
	canBeInstalled := PluginInstallReports{}
	canBeUpdated := PluginInstallReports{}

	for _, report := range reports {
		report.IsUpdateReport = isUpdateReport
		if !report.Skipped {
			installedOrUpdated = append(installedOrUpdated, report)
		} else if report.SkipReason == constants.PluginNotInstalled {
			canBeInstalled = append(canBeInstalled, report)
		} else if report.SkipReason == constants.PluginAlreadyInstalled {
			canBeUpdated = append(canBeUpdated, report)
		}
	}

	sort.Stable(reports)

	if len(installedOrUpdated) > 0 {
		fmt.Println()
		asString := []string{}
		for _, report := range installedOrUpdated {
			asString = append(asString, report.installString())
		}
		fmt.Println(strings.Join(asString, "\n\n"))
	}

	if len(installedOrUpdated) < len(reports) {
		skipCount := len(reports) - len(installedOrUpdated)
		asString := []string{}
		for _, report := range reports {
			if report.Skipped && report.SkipReason == constants.PluginNotInstalled {
				asString = append(asString, report.skipString())
			}
		}

		if (len(canBeInstalled) + len(canBeUpdated)) > 0 {
			fmt.Printf(
				"\nSkipped the following %s:\n\n%s\n",
				utils.Pluralize("plugin", skipCount),
				strings.Join(asString, "\n\n"),
			)
		}

		if len(canBeInstalled) > 0 {
			asString := []string{}
			for _, r := range canBeInstalled {
				asString = append(asString, r.Plugin)
			}
			fmt.Println()
			fmt.Printf(
				"To install %s which %s not installed, please run %s\n",
				utils.Pluralize("plugin", len(canBeInstalled)),
				utils.Pluralize("is", len(canBeInstalled)),
				constants.Bold(fmt.Sprintf(
					"steampipe plugin install %s",
					strings.Join(asString, " "),
				)),
			)
		}
		if len(canBeUpdated) > 0 {
			asString := []string{}
			for _, r := range canBeUpdated {
				asString = append(asString, r.Plugin)
			}
			fmt.Println()
			fmt.Printf(
				"To update %s which %s already installed, please run %s\n",
				utils.Pluralize("plugin", len(canBeUpdated)),
				utils.Pluralize("is", len(canBeUpdated)),
				constants.Bold(fmt.Sprintf(
					"steampipe plugin update %s",
					strings.Join(asString, " "),
				)),
			)
		}
	}
}
