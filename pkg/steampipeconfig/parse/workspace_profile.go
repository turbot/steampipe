package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func DecodeWorkspaceProfile(block *hcl.Block, parseContext ParseContext) (*modconfig.WorkspaceProfile, *decodeResult) {
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
