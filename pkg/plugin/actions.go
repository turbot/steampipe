package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

const (
	DefaultImageTag     = "latest"
	DefaultImageRepoURL = "us-docker.pkg.dev/steampipe/plugin"
	DefaultImageOrg     = "turbot"
)

// Remove removes an installed plugin
func Remove(ctx context.Context, image string, pluginConnections map[string][]modconfig.Connection) (*display.PluginRemoveReport, error) {
	statushooks.SetStatus(ctx, fmt.Sprintf("Removing plugin %s", image))
	defer statushooks.Done(ctx)

	imageRef := ociinstaller.NewSteampipeImageRef(image)
	fullPluginName := imageRef.DisplayImageRef()

	// are any connections using this plugin???
	conns := pluginConnections[fullPluginName]

	installedTo := filepath.Join(filepaths.EnsurePluginDir(), filepath.FromSlash(fullPluginName))
	_, err := os.Stat(installedTo)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin '%s' not found", image)
	}
	// remove from file system
	err = os.RemoveAll(installedTo)
	if err != nil {
		return nil, err
	}

	// update the version file
	v, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		return nil, err
	}
	delete(v.Plugins, fullPluginName)
	err = v.Save()

	return &display.PluginRemoveReport{Connections: conns, Image: imageRef}, err
}

// Exists looks up the version file and reports whether a plugin is already installed
func Exists(plugin string) (bool, error) {
	versionData, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		return false, err
	}

	imageRef := ociinstaller.NewSteampipeImageRef(plugin)

	// lookup in the version data
	_, found := versionData.Plugins[imageRef.DisplayImageRef()]
	return found, nil
}

// Install installs a plugin in the local file system
func Install(ctx context.Context, plugin string, sub chan struct{}) (*ociinstaller.SteampipeImage, error) {
	image, err := ociinstaller.InstallPlugin(ctx, plugin, sub)
	return image, err
}

// PluginListItem is a struct representing an item in the list of plugins
type PluginListItem struct {
	Name        string
	Version     string
	Connections []string
}

// List returns all installed plugins
func List(pluginConnectionMap map[string][]modconfig.Connection) ([]PluginListItem, error) {
	var items []PluginListItem

	var installedPlugins []string

	filepath.Walk(filepaths.EnsurePluginDir(), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".plugin") {
			rel, err := filepath.Rel(filepaths.EnsurePluginDir(), filepath.Dir(path))
			if err != nil {
				return err
			}
			installedPlugins = append(installedPlugins, rel)
		}
		return nil
	})

	v, err := versionfile.LoadPluginVersionFile()
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
			item.Connections = func() []string {
				// extract only the connection names
				conNames := []string{}
				for _, y := range pluginConnectionMap[plugin] {
					conNames = append(conNames, y.Name)
				}
				return conNames
			}()
		}
		items = append(items, item)
	}

	return items, nil
}
