package connection_config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/constants"
)

const maxSchemaNameLength = 63

// schemas in postgres are limited to 63 chars - the FQN may be longer than this, in which case trim the length
// and add a hash to the end to make unique
func PluginFQNToSchemaName(pluginFQN string) string {
	if len(pluginFQN) < maxSchemaNameLength {
		return pluginFQN
	}

	schemaName := trimSchemaName(pluginFQN) + fmt.Sprintf("-%x", utils.StringHash(pluginFQN))
	return schemaName
}

func trimSchemaName(pluginFQN string) string {
	if len(pluginFQN) < maxSchemaNameLength {
		return pluginFQN
	}

	return pluginFQN[:maxSchemaNameLength-9]
}

func GetPluginPath(remoteSchema string) (string, error) {
	log.Printf("[DEBUG] GetPluginPath for remote schema %s\n", remoteSchema)
	// the fully qualified name of the plugin is the relative path of the folder containing the plugin
	// calculate absolute folder path
	pluginFolder := filepath.Join(constants.PluginDir(), remoteSchema)

	// if the plugin folder is missing, it is possible the plugin path was truncated to create a schema name
	// - so search for a folder which when truncated would match the schema
	if _, err := os.Stat(pluginFolder); os.IsNotExist(err) {
		log.Printf("[DEBUG] plugin path %s not found - searching for folder using hashed name\n", pluginFolder)
		if pluginFolder, err = findPluginFolder(remoteSchema); err != nil {
			return "", err
		} else if pluginFolder == "" {
			return "", fmt.Errorf("no plugin installed matching schema %s", remoteSchema)
		}
	}

	// there should be just 1 file with extension pluginExtension ("fdw")
	entries, err := ioutil.ReadDir(pluginFolder)
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

// search for a folder which when hashed would match the schema
func findPluginFolder(remoteSchema string) (string, error) {
	pluginDir := constants.PluginDir()

	// first try searching by prefix - trim the schema name
	globPattern := filepath.Join(pluginDir, trimSchemaName(remoteSchema)) + "*"
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return "", err
	} else if len(matches) == 1 {
		return matches[0], nil
	}

	for _, match := range matches {
		// // get the relative path to this mat fromn the plugin folder
		folderRelativePath, err := filepath.Rel(pluginDir, match)
		if err != nil {
			// do not fail on error here
			continue
		}
		hashedName := PluginFQNToSchemaName(folderRelativePath)
		if hashedName == remoteSchema {
			log.Printf("[DEBUG] folder %s matches %s\n", folderRelativePath, remoteSchema)
			return filepath.Join(pluginDir, folderRelativePath), nil
		}
	}

	return "", nil

}
