package display

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/utils"
)

const ALREADY_INSTALLED = "Already installed"
const LATEST_ALREADY_INSTALLED = "Latest already installed"
const NOT_INSTALLED = "Not installed"

type InstallReport struct {
	Skipped        bool
	Plugin         string
	SkipReason     string
	DocURL         string
	Message        string
	Version        string
	IsUpdateReport bool
}

func (i *InstallReport) skipString() string {
	ref := ociinstaller.NewSteampipeImageRef(i.Plugin)
	_, name, stream := ref.GetOrgNameAndStream()

	return fmt.Sprintf("Plugin:   %s\nReason:   %s", fmt.Sprintf("%s@%s", name, stream), i.SkipReason)
}

func (i *InstallReport) installString() string {
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
		if len(i.Message) > 0 {
			thisReport = append(
				thisReport,
				fmt.Sprintf("Message:          %s", i.Message),
			)
		}
	}

	return strings.Join(thisReport, "\n")
}

func (i *InstallReport) String() string {
	if !i.Skipped {
		return i.installString()
	} else {
		return i.skipString()
	}
}

// Prints out the installation reports onto the console
func PrintInstallReports(reports []InstallReport, isUpdateReport bool) {
	installedOrUpdated := []InstallReport{}
	canBeInstalled := []InstallReport{}
	canBeUpdated := []InstallReport{}

	for _, report := range reports {
		report.IsUpdateReport = isUpdateReport
		if !report.Skipped {
			installedOrUpdated = append(installedOrUpdated, report)
		} else if report.SkipReason == NOT_INSTALLED {
			canBeInstalled = append(canBeInstalled, report)
		} else if report.SkipReason == ALREADY_INSTALLED {
			canBeUpdated = append(canBeUpdated, report)
		}
	}

	if len(installedOrUpdated) > 0 {
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
			if report.Skipped {
				asString = append(asString, report.skipString())
			}
		}
		// some have skipped
		if len(installedOrUpdated) > 0 {
			fmt.Println()
		}
		fmt.Printf(
			"Skipped the following %s:\n\n%s\n",
			utils.Pluralize("plugin", skipCount),
			strings.Join(asString, "\n\n"),
		)

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
