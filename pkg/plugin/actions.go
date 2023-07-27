package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/ociinstaller/versionfile"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/sperr"
)

const (
	DefaultImageTag     = "latest"
	DefaultImageRepoURL = "us-docker.pkg.dev/steampipe/plugin"
	DefaultImageOrg     = "turbot"
)

// Remove removes an installed plugin
func Remove(ctx context.Context, image string, pluginConnections map[string][]*modconfig.Connection) (*steampipeconfig.PluginRemoveReport, error) {
	statushooks.SetStatus(ctx, fmt.Sprintf("Removing plugin %s", image))

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

	return &steampipeconfig.PluginRemoveReport{Connections: conns, Image: imageRef}, err
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
	Version     *modconfig.PluginVersionString
	Connections []string
}

// List returns all installed plugins
func List(pluginConnectionMap map[string][]*modconfig.Connection) ([]PluginListItem, error) {
	var items []PluginListItem

	v, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		return nil, err
	}

	pluginVersions := v.Plugins

	pluginBinaries, err := files.ListFiles(filepaths.EnsurePluginDir(), &files.ListOptions{
		Include: []string{"**/*.plugin"},
		Flags:   files.AllRecursive,
	})
	if err != nil {
		return nil, err
	}

	// we have the plugin binary paths
	for _, pluginBinary := range pluginBinaries {
		parent := filepath.Dir(pluginBinary)
		fullPluginName, err := filepath.Rel(filepaths.EnsurePluginDir(), parent)
		if err != nil {
			return nil, err
		}
		// for local plugin
		item := PluginListItem{
			Name:    fullPluginName,
			Version: modconfig.LocalPluginVersionString(),
		}
		// check if this plugin is recorded in plugin versions
		installation, found := pluginVersions[fullPluginName]
		if found {
			// if not a local plugin, get the semver version
			if !detectLocalPlugin(installation, pluginBinary) {
				item.Version, err = modconfig.NewPluginVersionString(installation.Version)
				if err != nil {
					return nil, sperr.WrapWithMessage(err, "could not evaluate plugin version %s", installation.Version)
				}
			}

			if pluginConnectionMap != nil {
				// extract only the connection names
				var connectionNames []string
				for _, connection := range pluginConnectionMap[fullPluginName] {
					connectionName := connection.Name
					if connection.ImportDisabled() {
						connectionName = fmt.Sprintf("%s(disabled)", connectionName)
					}
					connectionNames = append(connectionNames, connectionName)
				}
				item.Connections = connectionNames
			}

			items = append(items, item)
		}
	}

	return items, nil
}

// detectLocalPlugin returns true if the modTime of the `pluginBinary` is after the installation date as recorded in the installation data
// this may happen when a plugin is installed from the registry, but is then compiled from source
func detectLocalPlugin(installation *versionfile.InstalledVersion, pluginBinary string) bool {
	installDate, err := time.Parse(time.RFC3339, installation.InstallDate)
	if err != nil {
		log.Printf("[WARN] could not parse install date for %s: %s", installation.Name, installation.InstallDate)
		return false
	}

	// truncate to second
	// otherwise, comparisons may get skewed because of the
	// underlying monotonic clock
	installDate = installDate.Truncate(time.Second)

	// get the modtime of the plugin binary
	stat, err := os.Lstat(pluginBinary)
	if err != nil {
		log.Printf("[WARN] could not parse install date for %s: %s", installation.Name, installation.InstallDate)
		return false
	}
	modTime := stat.ModTime().
		// truncate to second
		// otherwise, comparisons may get skewed because of the
		// underlying monotonic clock
		Truncate(time.Second)

	return installDate.Before(modTime)
}
