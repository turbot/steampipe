package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/ociinstaller/versionfile"
)

const (
	DefaultImageTag     = "latest"
	DefaultImageRepoURL = "us-docker.pkg.dev/steampipe/plugin"
	DefaultImageOrg     = "turbot"
)

// Remove removes an installed plugin
func Remove(image string, pluginConnections map[string][]string) error {
	spinner := display.ShowSpinner(fmt.Sprintf("Removing plugin %s", image))
	defer display.StopSpinner(spinner)

	fullPluginName := ociinstaller.NewSteampipeImageRef(image).DisplayImageRef()

	// are any connections using this plugin???
	conns, found := pluginConnections[fullPluginName]
	if found {
		return fmt.Errorf("there are active connections using it: '%s'", strings.Join(conns, ","))
	}

	installedTo := filepath.Join(constants.PluginDir(), filepath.FromSlash(fullPluginName))
	_, err := os.Stat(installedTo)
	if os.IsNotExist(err) {
		return fmt.Errorf("plugin '%s' not found", image)
	}
	// remove from file system
	err = os.RemoveAll(installedTo)
	if err != nil {
		return err
	}

	// update the version file
	v, err := versionfile.LoadForPlugin()
	if err != nil {
		return err
	}
	delete(v.Plugins, fullPluginName)
	return v.SaveForPlugin()
}

// Exists looks up the version file and reports whether a plugin is already installed
func Exists(plugin string) (bool, error) {
	versionData, err := versionfile.LoadForPlugin()
	if err != nil {
		return false, err
	}

	imageRef := ociinstaller.NewSteampipeImageRef(plugin)

	// lookup in the version data
	_, found := versionData.Plugins[imageRef.DisplayImageRef()]
	return found, nil
}

// Install installs a plugin in the local file system
func Install(plugin string) (*ociinstaller.SteampipeImage, error) {
	image, err := ociinstaller.InstallPlugin(plugin)
	return image, err
}

// PluginListItem is a struct representing an item in the list of plugins
type PluginListItem struct {
	Name        string
	Version     string
	Connections []string
}

// List returns all installed plugins
func List(pluginConnectionMap map[string][]string) ([]PluginListItem, error) {
	var items []PluginListItem

	var installedPlugins []string

	filepath.Walk(constants.PluginDir(), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".plugin") {
			rel, err := filepath.Rel(constants.PluginDir(), filepath.Dir(path))
			if err != nil {
				return err
			}
			installedPlugins = append(installedPlugins, rel)
		}
		return nil
	})

	v, err := versionfile.LoadForPlugin()
	if err != nil {
		return nil, err
	}

	pluginVersions := v.Plugins

	for _, plugin := range installedPlugins {
		version := "local"
		pluginDetails, found := pluginVersions[plugin]
		if found {
			version = pluginDetails.Version
		}
		item := PluginListItem{
			Name:    plugin,
			Version: version,
		}
		if pluginConnectionMap != nil {
			item.Connections = pluginConnectionMap[plugin]
		}
		items = append(items, item)
	}

	return items, nil
}
