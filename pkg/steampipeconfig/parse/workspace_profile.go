package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
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

	parseCtx := NewWorkspaceProfileParseContext(workspaceProfilePath)
	parseCtx.SetDecodeContent(content, fileData)

	// build parse context
	return parseWorkspaceProfiles(parseCtx)

}
func parseWorkspaceProfiles(parseCtx *WorkspaceProfileParseContext) (map[string]*modconfig.WorkspaceProfile, error) {

	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		_, diags := decodeWorkspaceProfiles(parseCtx)
		if diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to decode all mod hcl files", diags)
		}

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			return nil, fmt.Errorf("failed to resolve mod dependencies after %d attempts\nDependencies:\n%s", attempts+1, str)
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}

	return parseCtx.workspaceProfiles, nil

}

func decodeWorkspaceProfiles(parseCtx *WorkspaceProfileParseContext) (map[string]*modconfig.WorkspaceProfile, hcl.Diagnostics) {
	profileMap := map[string]*modconfig.WorkspaceProfile{}

	var diags hcl.Diagnostics
	blocksToDecode, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return nil, diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	for _, block := range blocksToDecode {
		workspaceProfile, res := decodeWorkspaceProfile(block, parseCtx)

		if res.Success() {
			// success - add to map
			profileMap[workspaceProfile.ProfileName] = workspaceProfile
		}
	}
	return profileMap, diags
}

// func DecodeWorkspaceProfiles(content *hcl.BodyContent, workspaceProfilePath string) map[string]*modconfig.WorkspaceProfile {
// }
func decodeWorkspaceProfile(block *hcl.Block, parseCtx *WorkspaceProfileParseContext) (*modconfig.WorkspaceProfile, *decodeResult) {
	res := newDecodeResult()
	// get shell resource
	resource := modconfig.NewWorkspaceProfile(block)

	diags := gohcl.DecodeBody(block.Body, parseCtx.EvalCtx, resource)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
	}

	handleWorkspaceProfileDecodeResult(resource, res, block, parseCtx)
	return resource, res
}

func handleWorkspaceProfileDecodeResult(resource *modconfig.WorkspaceProfile, res *decodeResult, block *hcl.Block, parseCtx *WorkspaceProfileParseContext) {
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded()
		res.addDiags(moreDiags)

		moreDiags = parseCtx.AddResource(resource)
		res.addDiags(moreDiags)
		return
	}

	// failure :(
	if len(res.Depends) > 0 {
		moreDiags := parseCtx.AddDependencies(block, resource.Name(), res.Depends)
		res.addDiags(moreDiags)
	}
}
