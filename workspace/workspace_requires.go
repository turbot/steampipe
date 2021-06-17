package workspace

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/steampipe/constants"

	version "github.com/hashicorp/go-version"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/plugin"
	"github.com/turbot/steampipe/utils"
)

func (w *Workspace) CheckRequiredPluginsInstalled() error {

	// get the list of all installed plugins
	installedPlugins, err := w.getInstalledPlugins()
	if err != nil {
		return err
	}

	// get the list of all the required plugins
	requiredPlugins := w.getRequiredPlugins()

	var versionFailures []requiredPluginVersion

	for name, _ := range requiredPlugins {
		if _, found := installedPlugins[name]; found {
			if installedPlugins[name].LessThan(requiredPlugins[name]) {
				versionFailures = append(versionFailures,
					requiredPluginVersion{
						plugin:           name,
						installedVersion: installedPlugins[name].String(),
						requiredVersion:  requiredPlugins[name].String(),
					})
			}
		} else {
			versionFailures = append(versionFailures,
				requiredPluginVersion{
					plugin:           name,
					installedVersion: "none",
					requiredVersion:  requiredPlugins[name].String(),
				})
		}

	}
	if len(versionFailures) > 0 {
		return errors.New(pluginVersionError(versionFailures))

	}
	return nil
}

func (w *Workspace) getRequiredPlugins() map[string]*version.Version {
	if w.Mod.Requires != nil {
		requiredPluginVersions := w.Mod.Requires.Plugins
		requiredVersion := make(map[string]*version.Version)
		for _, pluginVersion := range requiredPluginVersions {
			requiredVersion[pluginVersion.ShortName()] = pluginVersion.ParsedVersion
		}
		return requiredVersion
	}
	return nil
}

func (w *Workspace) getInstalledPlugins() (map[string]*version.Version, error) {
	installedPlugins := make(map[string]*version.Version)
	installedPluginsData, _ := plugin.List(nil)
	for _, plugin := range installedPluginsData {
		org, name, _ := ociinstaller.NewSteampipeImageRef(plugin.Name).GetOrgNameAndStream()
		semverVersion, err := version.NewVersion(plugin.Version)
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

func pluginVersionError(reqs []requiredPluginVersion) string {
	var notificationLines = []string{
		fmt.Sprintf("%d mod plugin %s are not satisfied. ", len(reqs), utils.Pluralize("requirement", len(reqs))),
		"",
	}
	longestNameLength := 0
	for _, report := range reqs {
		thisName := report.plugin
		if len(thisName) > longestNameLength {
			longestNameLength = len(thisName)
		}
	}

	// sort alphabetically
	sort.Slice(reqs, func(i, j int) bool {
		return reqs[i].plugin < reqs[j].plugin
	})

	// build first part of string
	// recheck longest names
	longestVersionLength := 0
	var pluginVersions = make([]string, len(reqs))
	for i, req := range reqs {
		format := fmt.Sprintf("  %%-%ds  %%-2s", longestNameLength)
		pluginVersions[i] = fmt.Sprintf(
			format,
			req.plugin,
			req.installedVersion,
		)

		if len(pluginVersions[i]) > longestVersionLength {
			longestVersionLength = len(pluginVersions[i])
		}
	}

	for i, req := range reqs {
		format := fmt.Sprintf("%%-%ds  →  %%2s", longestVersionLength)
		notificationLines = append(notificationLines, fmt.Sprintf(
			format,
			pluginVersions[i],
			constants.Bold(req.requiredVersion),
		))
	}

	// add blank line (hack - bold the empty string to force it to print blank line as part of error)
	notificationLines = append(notificationLines, fmt.Sprintf("%s", constants.Bold("")))

	return strings.Join(notificationLines, "\n")
}
