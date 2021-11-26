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
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

const (
	DefaultImageTag     = "latest"
	DefaultImageRepoURL = "us-docker.pkg.dev/steampipe/plugin"
	DefaultImageOrg     = "turbot"
)

// Remove removes an installed plugin
func Remove(image string, pluginConnections map[string][]modconfig.Connection) error {
	spinner := display.ShowSpinner(fmt.Sprintf("Removing plugin %s", image))
	defer display.StopSpinner(spinner)

	fullPluginName := ociinstaller.NewSteampipeImageRef(image).DisplayImageRef()

	// are any connections using this plugin???
	conns := pluginConnections[fullPluginName]

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
	v, err := versionfile.LoadPluginVersionFile()
	if err != nil {
		return err
	}
	delete(v.Plugins, fullPluginName)
	err = v.Save()

	// store the filenames of the config files, that have the connections
	var files = map[int]string{}
	if len(conns) > 0 {
		for i, con := range conns {
			files[i] = con.DeclRange.Filename
		}
	}
	connFiles := Unique(files)

	if len(conns) > 0 {
		display.StopSpinner(spinner)
		str := []string{fmt.Sprintf("\nNote: the following %s have steampipe connections using the '%s' plugin:\n", utils.Pluralize("file", len(connFiles)), image)}
		for _, conn := range conns {
			str = append(
				str,
				fmt.Sprintf(
					"\t* file: %s \n \t  connection: '%s' (line %d)",
					conn.DeclRange.Filename,
					conn.Name,
					conn.DeclRange.Start.Line,
				),
			)
		}
		str = append(str, "\nPlease remove them to continue using steampipe")
		fmt.Println(strings.Join(str, "\n"))
		fmt.Println()
	}

	return err
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
func List(pluginConnectionMap map[string][]modconfig.Connection) ([]PluginListItem, error) {
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

// function that returns an unique map of strings
func Unique(m map[int]string) map[int]string {
	n := make(map[int]string, len(m))
	ref := make(map[string]bool, len(m))
	for k, v := range m {
		if _, ok := ref[v]; !ok {
			ref[v] = true
			n[k] = v
		}
	}
	return n
}
