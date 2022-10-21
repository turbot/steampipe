package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func LoadWorkspaceProfiles(workspaceProfilePath string) (profileMap map[string]*modconfig.WorkspaceProfile, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
		// be sure to return the default
		if profileMap != nil && profileMap["default"] == nil {
			profileMap["default"] = &modconfig.WorkspaceProfile{ProfileName: "default"}
		}
	}()

	// create profile map to populate
	profileMap = map[string]*modconfig.WorkspaceProfile{}

	configPaths, err := filehelpers.ListFiles(workspaceProfilePath, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
	})
	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		return profileMap, nil
	}

	fileData, diags := LoadFileData(configPaths...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, diags := body.Content(WorkspaceProfileListBlockSchema)
	if diags.HasErrors() {

		return nil, plugin.DiagsToError("Failed to load workspace profiles", diags)
	}

	// build parse context
	return parseWorkspaceProfiles(content, workspaceProfilePath)

}
func parseWorkspaceProfiles(content *hcl.BodyContent, workspaceProfilePath string) (map[string]*modconfig.WorkspaceProfile, error) {
	parseContext := NewParseContext(workspaceProfilePath)

	profileMap := map[string]*modconfig.WorkspaceProfile{}
	for _, block := range content.Blocks {

		workspaceProfile, res := decodeWorkspaceProfile(block, parseContext)

		if res.Success() {
			// success - add to map
			profileMap[workspaceProfile.ProfileName] = workspaceProfile
		}
		// TODO handle failure and dependencies
	}
	return profileMap, nil

}

// func DecodeWorkspaceProfiles(content *hcl.BodyContent, workspaceProfilePath string) map[string]*modconfig.WorkspaceProfile {
// }
func decodeWorkspaceProfile(block *hcl.Block, parseContext ParseContext) (*modconfig.WorkspaceProfile, *decodeResult) {
	res := newDecodeResult()
	// get shell resource
	resource := modconfig.NewWorkspaceProfile(block)

	diags := gohcl.DecodeBody(block.Body, parseContext.EvalCtx, resource)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
	}
	return resource, res

	//WorkspaceProfileContent, rest, diags := block.Body.PartialContent(WorkspaceProfileBlockSchema)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	//var ctx  = hcl.EvalContext
	//
	//// get WorkspaceProfile name
	//resource := modconfig.NewWorkspaceProfile(block)
	//
	//res := newDecodeResult()
	//
	//diags = gohcl.DecodeBody(block.Body, runCtx.EvalCtx, resource)
	//if len(diags) > 0 {
	//	res.handleDecodeDiags(diags)
	//}
	//return resource, res
	//
	//var pluginName string
	//diags = gohcl.DecodeExpression(WorkspaceProfileContent.Attributes["plugin"].Expr, nil, &pluginName)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	//
	//if strings.HasPrefix(pluginName, "local/") {
	//	WorkspaceProfile.Plugin = pluginName
	//} else {
	//	WorkspaceProfile.Plugin = ociinstaller.NewSteampipeImageRef(pluginName).DisplayImageRef()
	//}
	//WorkspaceProfile.PluginShortName = pluginName
	//
	//if WorkspaceProfileContent.Attributes["type"] != nil {
	//	var WorkspaceProfileType string
	//	diags = gohcl.DecodeExpression(WorkspaceProfileContent.Attributes["type"].Expr, nil, &WorkspaceProfileType)
	//	if diags.HasErrors() {
	//		return nil, diags
	//	}
	//	WorkspaceProfile.Type = WorkspaceProfileType
	//}
	//if WorkspaceProfileContent.Attributes["WorkspaceProfiles"] != nil {
	//	var WorkspaceProfiles []string
	//	diags = gohcl.DecodeExpression(WorkspaceProfileContent.Attributes["WorkspaceProfiles"].Expr, nil, &WorkspaceProfiles)
	//	if diags.HasErrors() {
	//		return nil, diags
	//	}
	//	WorkspaceProfile.WorkspaceProfileNames = WorkspaceProfiles
	//}
	//
	//// check for nested options
	//for _, WorkspaceProfileBlock := range WorkspaceProfileContent.Blocks {
	//	switch WorkspaceProfileBlock.Type {
	//	case "options":
	//		// if we already found settings, fail
	//		opts, moreDiags := DecodeOptions(WorkspaceProfileBlock)
	//		if moreDiags.HasErrors() {
	//			diags = append(diags, moreDiags...)
	//			break
	//		}
	//		moreDiags = WorkspaceProfile.SetOptions(opts, WorkspaceProfileBlock)
	//		if moreDiags.HasErrors() {
	//			diags = append(diags, moreDiags...)
	//		}
	//
	//	default:
	//		// this can probably never happen
	//		diags = append(diags, &hcl.Diagnostic{
	//			Severity: hcl.DiagError,
	//			Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for WorkspaceProfiles", WorkspaceProfileBlock.Type),
	//			Subject:  &WorkspaceProfileBlock.DefRange,
	//		})
	//	}
	//}
	//// convert the remaining config to a hcl string to pass to the plugin
	//config, moreDiags := pluginWorkspaceProfileConfigToHclString(rest, WorkspaceProfileContent)
	//if moreDiags.HasErrors() {
	//	diags = append(diags, moreDiags...)
	//} else {
	//	WorkspaceProfile.Config = config
	//}
	//
	//return WorkspaceProfile, diags
}
