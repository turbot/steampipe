package filepaths

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

func GetPluginPath(plugin, pluginShortName string) (string, error) {
	remoteSchema := plugin
	// the fully qualified name of the plugin is the relative path of the folder containing the plugin
	// calculate absolute folder path
	pluginFolder := filepath.Join(EnsurePluginDir(), remoteSchema)

	// if the plugin folder is missing, it is possible the plugin path was truncated to create a schema name
	// - so search for a folder which when truncated would match the schema
	if _, err := os.Stat(pluginFolder); os.IsNotExist(err) {
		log.Printf("[TRACE] plugin path %s not found - searching for folder using hashed name\n", pluginFolder)
		if pluginFolder, err = FindPluginFolder(remoteSchema); err != nil {
			return "", err
		} else if pluginFolder == "" {
			return "", fmt.Errorf("no plugin installed matching %s", pluginShortName)
		}
	}

	// there should be just 1 file with extension pluginExtension (".plugin")
	entries, err := os.ReadDir(pluginFolder)
	if err != nil {
		return "", fmt.Errorf("failed to load plugin %s: %v", remoteSchema, err)
	}
	matches := []string{}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == constants.PluginExtension {
			matches = append(matches, entry.Name())
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("plugin folder %s should contain a single plugin file. %d plugins were found ", pluginFolder, len(matches))
	}

	return filepath.Join(pluginFolder, matches[0]), nil
}

// FindPluginFolder searches for a folder which when hashed would match the schema
func FindPluginFolder(remoteSchema string) (string, error) {
	pluginDir := EnsurePluginDir()

	// first try searching by prefix - trim the schema name
	globPattern := filepath.Join(pluginDir, utils.TrimSchemaName(remoteSchema)) + "*"
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return "", err
	} else if len(matches) == 1 {
		return matches[0], nil
	}

	for _, match := range matches {
		// get the relative path to this match from the plugin folder
		folderRelativePath, err := filepath.Rel(pluginDir, match)
		if err != nil {
			// do not fail on error here
			continue
		}
		hashedName := utils.PluginFQNToSchemaName(folderRelativePath)
		if hashedName == remoteSchema {
			return filepath.Join(pluginDir, folderRelativePath), nil
		}
	}

	return "", nil
}
