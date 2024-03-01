package constants

import "github.com/turbot/go-kit/helpers"

var ModDataExtensions = []string{".sp"}
var VariablesExtensions = []string{".spvars"}
var AutoVariablesExtensions = []string{".auto.spvars"}

const (
	PluginExtension      = ".plugin"
	ConfigExtension      = ".spc"
	SqlExtension         = ".sql"
	JsonExtension        = ".json"
	TextExtension        = ".txt"
	SnapshotExtension    = ".sps"
	TokenExtension       = ".tptt"
	LegacyTokenExtension = ".sptt"
)

var YamlExtensions = []string{".yml", ".yaml"}

var ConnectionConfigExtensions = append(YamlExtensions, ConfigExtension, JsonExtension)

func IsYamlExtension(ext string) bool {
	return helpers.StringSliceContains(YamlExtensions, ext)
}
