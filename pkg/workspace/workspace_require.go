package workspace

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"

	"github.com/Masterminds/semver"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/plugin"
	"github.com/turbot/steampipe/pkg/utils"
)

func (w *Workspace) CheckRequiredPluginsInstalled() error {
	// get the list of all installed plugins
	installedPlugins, err := w.getInstalledPlugins()
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
			req.SetInstalledVersion(installedVersion)

			if installedPlugins[name].LessThan(requiredVersion) {
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

func (w *Workspace) ValidateSteampipeVersion() error {
	return w.Mod.ValidateSteampipeVersion()
}

func (w *Workspace) getRequiredPlugins() map[string]*semver.Version {
	if w.Mod.Require != nil {
		requiredPluginVersions := w.Mod.Require.Plugins
		requiredVersion := make(versionmap.VersionMap)
		for _, pluginVersion := range requiredPluginVersions {
			requiredVersion[pluginVersion.ShortName()] = pluginVersion.Version
		}
		return requiredVersion
	}
	return nil
}

func (w *Workspace) getInstalledPlugins() (versionmap.VersionMap, error) {
	installedPlugins := make(versionmap.VersionMap)
	installedPluginsData, _ := plugin.List(nil)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewSteampipeImageRef(plugin.Name).GetOrgNameAndStream()
		semverVersion, err := semver.NewVersion(plugin.Version)
		if err != nil {
			continue
		}
		pluginShortName := fmt.Sprintf("%s/%s", org, name)
		installedPlugins[pluginShortName] = semverVersion
	}
	return installedPlugins, nil
}

type requiredPluginVersion struct {
	plugin           string
	requiredVersion  string
	installedVersion string
}

func (v *requiredPluginVersion) SetRequiredVersion(requiredVersion *semver.Version) {
	requiredVersionString := requiredVersion.String()
	// if no required version was specified, the version will be 0.0.0
	if requiredVersionString == "0.0.0" {
		v.requiredVersion = "latest"
	} else {
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
		_, plugin, _ := strings.Cut(req.plugin, "/")

		// check if plugin needs to be installed/updated
		if strings.Contains(notInstalledStrings[i], "none") {
			notificationLines = append(notificationLines, fmt.Sprintf(
				"  steampipe plugin install %s", plugin,
			))
		} else {
			notificationLines = append(notificationLines, fmt.Sprintf(
				"  steampipe plugin update %s", plugin,
			))
		}
	}

	// add blank line (hack - bold the empty string to force it to print blank line as part of error)
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
