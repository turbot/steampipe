package workspace

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/ociinstaller"

	version "github.com/hashicorp/go-version"
	"github.com/turbot/steampipe/plugin"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func (w *Workspace) CheckRequiredPluginsInstalled() error {
	var errors []error

	// get the list of all installed plugins
	installedPlugins, err := w.getInstalledPlugins()
	if err != nil {
		return err
	}

	// get the list of all the required plugins
	requiredPlugins := w.getRequiredPlugins()

	for name, _ := range requiredPlugins {
		if _, found := installedPlugins[name]; found {
			if installedPlugins[name].LessThan(requiredPlugins[name]) {
				errors = append(errors, fmt.Errorf("plugin: '%s', required: %s, installed: %s", name, requiredPlugins[name], installedPlugins[name]))
			}
		} else {
			errors = append(errors, fmt.Errorf("plugin: '%s', required: %s, installed: none", name, requiredPlugins[name]))
		}

	}
	if len(errors) > 0 {

		var combinedError []string
		for _, err := range errors {
			combinedError = append(combinedError, err.Error())
		}
		//message := fmt.Sprintf()

		return fmt.Errorf(strings.Join(combinedError, "\n"))
	}
	return nil
}

func (w *Workspace) getRequiredPlugins() map[string]*version.Version {
	if w.Mod.Requires != nil {
		requiredPluginVersions := w.Mod.Requires.Plugins
		requiredVersion := make(map[string]*version.Version)
		for _, pluginVersion := range requiredPluginVersions {
			requiredVersion[pluginVersion.Name] = pluginVersion.ParsedVersion
		}
		return requiredVersion
	}
	return nil
}

func (w *Workspace) getInstalledPlugins() (map[string]*version.Version, error) {
	installedPlugins := make(map[string]*version.Version)
	installedPluginsData, _ := plugin.List(nil)
	for _, plugin := range installedPluginsData {
		_, name, _ := ociinstaller.NewSteampipeImageRef(plugin.Name).GetOrgNameAndStream()
		semverVersion, err := version.NewVersion(plugin.Version)
		if err != nil {
			continue
		}
		installedPlugins[name] = semverVersion
	}
	return installedPlugins, nil
}

// load all dependencies of workspace mod
// used to load all mods in a workspace, using the workspace manifest mod
func (w *Workspace) loadModDependencies(modsFolder string) (modconfig.ModMap, error) {
	var res = modconfig.ModMap{
		w.Mod.Name(): w.Mod,
	}
	if err := steampipeconfig.LoadModDependencies(w.Mod, modsFolder, res, false); err != nil {
		return nil, err
	}
	return res, nil
}
