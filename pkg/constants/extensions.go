package constants

import "github.com/turbot/go-kit/helpers"

const (
	PluginExtension        = ".plugin"
	ConfigExtension        = ".spc"
	SqlExtension           = ".sql"
	MarkdownExtension      = ".md"
	ModDataExtension       = ".sp"
	VariablesExtension     = ".spvars"
	AutoVariablesExtension = ".auto.spvars"
	JsonExtension          = ".json"
	CsvExtension           = ".csv"
	TextExtension          = ".txt"
	SnapshotExtension      = ".sps"
	TokenExtension         = ".sptt"
)

var YamlExtensions = []string{".yml", ".yaml"}

var ConnectionConfigExtensions = append(YamlExtensions, ConfigExtension, JsonExtension)

func IsYamlExtension(ext string) bool {
	return helpers.StringSliceContains(YamlExtensions, ext)
}
