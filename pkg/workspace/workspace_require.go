package workspace

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/plugin"
	"github.com/turbot/steampipe/pkg/utils"
)

func (w *Workspace) CheckRequiredPluginsInstalled() error {
	// get the list of all installed plugins
	installedPlugins, err := plugin.GetInstalledPlugins()
	if err != nil {
		return err
	}

	// get the list of all the required plugins
	requiredPlugins := w.getRequiredPlugins()

	var pluginsNotInstalled []requiredPluginVersion

	for name, requiredVersion := range requiredPlugins {
		var req = requiredPluginVersion{plugin: name}
		req.SetRequiredVersion(requiredVersion)

		if installedVersion, found := installedPlugins[name]; found {
			if installedVersion.IsLocal() {
				req.installedVersion = installedVersion.String()
				continue
			}
			smv := installedVersion.Semver()
			req.SetInstalledVersion(smv)

			if !requiredVersion.Check(smv) {
				pluginsNotInstalled = append(pluginsNotInstalled, req)
			}
		} else {
			req.installedVersion = "none"
			pluginsNotInstalled = append(pluginsNotInstalled, req)
		}

	}
	if len(pluginsNotInstalled) > 0 {
		return errors.New(pluginVersionError(pluginsNotInstalled))
	}

	return nil
}

func (w *Workspace) getRequiredPlugins() map[string]*semver.Constraints {
	if w.Mod.Require != nil {
		requiredPluginVersions := w.Mod.Require.Plugins
		requiredVersion := make(map[string]*semver.Constraints)
		for _, pluginVersion := range requiredPluginVersions {
			requiredVersion[pluginVersion.ShortName()] = pluginVersion.Constraint
		}
		return requiredVersion
	}
	return nil
}

type requiredPluginVersion struct {
	plugin           string
	requiredVersion  string
	installedVersion string
}

func (v *requiredPluginVersion) SetRequiredVersion(requiredVersion *semver.Constraints) {
	if requiredVersion == nil {
		v.requiredVersion = "*"
	} else {
		requiredVersionString := requiredVersion.String()
		v.requiredVersion = requiredVersionString
	}
}

func (v *requiredPluginVersion) SetInstalledVersion(installedVersion *semver.Version) {
	v.installedVersion = installedVersion.String()
}

func pluginVersionError(pluginsNotInstalled []requiredPluginVersion) string {
	failureCount := len(pluginsNotInstalled)
	var notificationLines = []string{
		fmt.Sprintf("%d mod plugin %s not satisfied. ", failureCount, utils.Pluralize("requirement", failureCount)),
		"",
	}
	longestNameLength := 0
	for _, report := range pluginsNotInstalled {
		thisName := report.plugin
		if len(thisName) > longestNameLength {
			longestNameLength = len(thisName)
		}
	}

	// sort alphabetically
	sort.Slice(pluginsNotInstalled, func(i, j int) bool {
		return pluginsNotInstalled[i].plugin < pluginsNotInstalled[j].plugin
	})

	// build first part of string
	// recheck longest names
	longestVersionLength := 0

	var notInstalledStrings = make([]string, len(pluginsNotInstalled))
	for i, req := range pluginsNotInstalled {
		format := fmt.Sprintf("  %%-%ds  %%-2s", longestNameLength)
		notInstalledStrings[i] = fmt.Sprintf(
			format,
			req.plugin,
			req.installedVersion,
		)

		if len(notInstalledStrings[i]) > longestVersionLength {
			longestVersionLength = len(notInstalledStrings[i])
		}
	}

	for i, req := range pluginsNotInstalled {
		format := fmt.Sprintf("%%-%ds  →  %%2s", longestVersionLength)
		notificationLines = append(notificationLines, fmt.Sprintf(
			format,
			notInstalledStrings[i],
			constants.Bold(req.requiredVersion),
		))
	}

	// add help message for missing plugins
	msg := fmt.Sprintf("\nPlease %s the %s with: \n", checkInstallOrUpdate(pluginsNotInstalled), utils.Pluralize("plugin", len(pluginsNotInstalled)))
	notificationLines = append(notificationLines, msg)

	for i, req := range pluginsNotInstalled {
		_, p, _ := strings.Cut(req.plugin, "/")

		// check if plugin needs to be installed/updated
		if strings.Contains(notInstalledStrings[i], "none") {
			notificationLines = append(notificationLines, fmt.Sprintf(
				"  steampipe plugin install %s", p,
			))
		} else {
			notificationLines = append(notificationLines, fmt.Sprintf(
				"  steampipe plugin update %s", p,
			))
		}
	}

	// add blank line (tactical - bold the empty string to force it to print blank line as part of error)
	notificationLines = append(notificationLines, fmt.Sprintf("%s", constants.Bold("")))

	return strings.Join(notificationLines, "\n")
}

// function to check whether the missing plugins require to be installed or updated, or both
func checkInstallOrUpdate(pluginsNotInstalled []requiredPluginVersion) string {
	var updateFlag, installFlag bool

	for _, req := range pluginsNotInstalled {
		if strings.Contains(req.installedVersion, "none") {
			installFlag = true
		} else {
			updateFlag = true
		}
	}

	if updateFlag {
		if installFlag {
			return "install/update"
		} else {
			return "update"
		}
	}
	return "install"
}
