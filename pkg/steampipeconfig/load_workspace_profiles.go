package steampipeconfig

import (
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
)

var GlobalWorkspaceProfile *modconfig.WorkspaceProfile

func LoadWorkspaceProfiles(configFolder string) (map[string]*modconfig.WorkspaceProfile, error) {
	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
	})

	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		return nil, nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.WorkspaceProfileListBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	profileMap := map[string]*modconfig.WorkspaceProfile{}
	// build parse context
	parseContext := parse.NewParseContext(configFolder)
	for _, block := range content.Blocks {

		workspaceProfile, res := parse.DecodeWorkspaceProfile(block, parseContext)
		if res.Success() {
			profileMap[workspaceProfile.Name] = workspaceProfile
		}
		//if moreDiags.HasErrors() {
		//	diags = append(diags, moreDiags...)
		//	continue
		//}
		//_, alreadyThere := steampipeConfig.Connections[connection.Name]
		//if alreadyThere {
		//	return fmt.Errorf("duplicate connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
		//}
		//if ok, errorMessage := schema.IsSchemaNameValid(connection.Name); !ok {
		//	return fmt.Errorf("invalid connection name: '%s' in '%s'. %s ", connection.Name, block.TypeRange.Filename, errorMessage)
		//}
		//steampipeConfig.Connections[connection.Name] = connection
	}

	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load config", diags)
	}

	// add in default if needed
	if _, ok := profileMap["default"]; !ok {
		// todo KAI populate default profile
		profileMap["default"] = &modconfig.WorkspaceProfile{}
	}

	return profileMap, nil
}
