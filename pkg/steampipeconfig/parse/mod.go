package parse

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func LoadModfile(modPath string) (*modconfig.Mod, error) {
	if !ModfileExists(modPath) {
		return nil, nil
	}

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: ContextFunctions(modPath),
		Variables: make(map[string]cty.Value),
	}

	mod, res := ParseModDefinition(modPath, evalCtx)
	if res.Diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load mod", res.Diags)
	}

	return mod, nil
}

// ParseModDefinition parses the modfile only
// it is expected the calling code will have verified the existence of the modfile by calling ModfileExists
// this is called before parsing the workspace to, for example, identify dependency mods
func ParseModDefinition(modPath string, evalCtx *hcl.EvalContext) (*modconfig.Mod, *DecodeResult) {
	res := newDecodeResult()

	// if there is no mod at this location, return error
	modFilePath := filepaths.ModFilePath(modPath)
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("no mod file found in %s", modPath),
		})
		return nil, res
	}

	fileData, diags := LoadFileData(modFilePath)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	body, diags := ParseHclFiles(fileData)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	workspaceContent, diags := body.Content(WorkspaceBlockSchema)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	block := hclhelpers.GetFirstBlockOfType(workspaceContent.Blocks, modconfig.BlockTypeMod)
	if block == nil {
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("no mod definition found in %s", modPath),
		})
		return nil, res
	}
	var defRange = block.DefRange
	if hclBody, ok := block.Body.(*hclsyntax.Body); ok {
		defRange = hclBody.SrcRange
	}
	mod := modconfig.NewMod(block.Labels[0], path.Dir(modFilePath), defRange)
	// set modFilePath
	mod.SetFilePath(modFilePath)

	mod, res = decodeMod(block, evalCtx, mod)
	if res.Diags.HasErrors() {
		return nil, res
	}

	// NOTE: IGNORE DEPENDENCY ERRORS

	// call decode callback
	diags = mod.OnDecoded(block, nil)
	res.addDiags(diags)

	return mod, res
}

// ParseMod parses all source hcl files for the mod path and associated resources, and returns the mod object
// NOTE: the mod definition has already been parsed (or a default created) and is in opts.RunCtx.RootMod
func ParseMod(fileData map[string][]byte, pseudoResources []modconfig.MappableResource, parseCtx *ModParseContext) (*modconfig.Mod, *modconfig.ErrorAndWarnings) {
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to load all mod source files", diags))
	}

	content, moreDiags := body.Content(WorkspaceBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to load mod", diags))
	}

	mod := parseCtx.CurrentMod
	if mod == nil {
		return nil, modconfig.NewErrorsAndWarning(fmt.Errorf("ParseMod called with no Current Mod set in ModParseContext"))
	}
	// get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources, err := loadMappableResourceNames(content)
	if err != nil {
		return nil, modconfig.NewErrorsAndWarning(err)
	}

	// if variables were passed in parsecontext, add to the mod
	if parseCtx.Variables != nil {
		for _, v := range parseCtx.Variables.RootVariables {
			if diags = mod.AddResource(v); diags.HasErrors() {
				return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to add resource to mod", diags))
			}
		}
	}

	// add pseudo resources to the mod
	addPseudoResourcesToMod(pseudoResources, hclResources, mod)

	// add the parsed content to the run context
	parseCtx.SetDecodeContent(content, fileData)

	// add the mod to the run context
	// - this it to ensure all pseudo resources get added and build the eval context with the variables we just added
	if diags = parseCtx.AddModResources(mod); diags.HasErrors() {
		return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to add mod to run context", diags))
	}

	// collect warnings as we parse
	var res = &modconfig.ErrorAndWarnings{}

	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		diags = decode(parseCtx)
		if diags.HasErrors() {
			return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to decode all mod hcl files", diags))
		}
		// now retrieve the warning strings
		res.AddWarning(plugin.DiagsToWarnings(diags)...)

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			return nil, modconfig.NewErrorsAndWarning(fmt.Errorf("failed to resolve dependencies for mod '%s' after %d attempts\nDependencies:\n%s", mod.FullName, attempts+1, str))
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}

	// now tell mod to build tree of resources
	res.Error = mod.BuildResourceTree(parseCtx.GetTopLevelDependencyMods())

	return mod, res
}
