package parse

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"sigs.k8s.io/yaml"
)

// LoadFileData builds a map of filepath to file data
func LoadFileData(paths ...string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var fileData = map[string][]byte{}

	for _, configPath := range paths {
		data, err := os.ReadFile(configPath)

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error()})
			continue
		}
		fileData[configPath] = data
	}
	return fileData, diags
}

// ParseHclFiles parses hcl file data and returns the hcl body object
func ParseHclFiles(fileData map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	var parsedConfigFiles []*hcl.File
	var diags hcl.Diagnostics
	parser := hclparse.NewParser()

	// build ordered list of files so that we parse in a repeatable order
	filePaths := buildOrderedFileNameList(fileData)

	for _, filePath := range filePaths {
		var file *hcl.File
		var moreDiags hcl.Diagnostics
		ext := filepath.Ext(filePath)
		if ext == constants.JsonExtension {
			file, moreDiags = json.ParseFile(filePath)
		} else if constants.IsYamlExtension(ext) {
			file, moreDiags = parseYamlFile(filePath)
		} else {
			data := fileData[filePath]
			file, moreDiags = parser.ParseHCL(data, filePath)
		}

		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
}

func buildOrderedFileNameList(fileData map[string][]byte) []string {
	filePaths := make([]string, len(fileData))
	idx := 0
	for filePath := range fileData {
		filePaths[idx] = filePath
		idx++
	}
	sort.Strings(filePaths)
	return filePaths
}

// ModfileExists returns whether a mod file exists at the specified path
func ModfileExists(modPath string) bool {
	modFilePath := filepath.Join(modPath, "mod.sp")
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		return false
	}
	return true
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

	// create a temporary runContext to decode the mod definition
	// note - this is not fully populated - the only properties which will be used are

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
	for _, v := range parseCtx.Variables {
		if diags = mod.AddResource(v); diags.HasErrors() {
			return nil, modconfig.NewErrorsAndWarning(plugin.DiagsToError("Failed to add resource to mod", diags))
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

// parse a yaml file into a hcl.File object
func parseYamlFile(filename string) (*hcl.File, hcl.Diagnostics) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to open file",
				Detail:   fmt.Sprintf("The file %q could not be opened.", filename),
			},
		}
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while reading it.", filename),
			},
		}
	}
	jsonData, err := yaml.YAMLToJSON(src)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read convert YAML to JSON",
				Detail:   fmt.Sprintf("The file %q was opened, but an error occured while converting it to JSON.", filename),
			},
		}
	}
	return json.Parse(jsonData, filename)
}

func addPseudoResourcesToMod(pseudoResources []modconfig.MappableResource, hclResources map[string]bool, mod *modconfig.Mod) {
	var duplicates []string
	for _, r := range pseudoResources {
		// is there a hcl resource with the same name as this pseudo resource - it takes precedence
		name := r.GetUnqualifiedName()
		if _, ok := hclResources[name]; ok {
			duplicates = append(duplicates, r.GetDeclRange().Filename)
			continue
		}
		// add pseudo resource to mod
		mod.AddResource(r.(modconfig.HclResource))
		// add to map of existing resources
		hclResources[name] = true
	}
	numDupes := len(duplicates)
	if numDupes > 0 {
		log.Printf("[TRACE] %d %s  not converted into resources as hcl resources of same name are defined: %v", numDupes, utils.Pluralize("file", numDupes), duplicates)
	}
}

// get names of all resources defined in hcl which may also be created as pseudo resources
// if we find a mod block, build a shell mod
func loadMappableResourceNames(content *hcl.BodyContent) (map[string]bool, error) {
	hclResources := make(map[string]bool)

	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {
		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			hclResources[name] = true
		}
	}
	return hclResources, nil
}

// ParseModResourceNames parses all source hcl files for the mod path and associated resources,
// and returns the resource names
func ParseModResourceNames(fileData map[string][]byte) (*modconfig.WorkspaceResources, error) {
	var resources = modconfig.NewWorkspaceResources()
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, moreDiags := body.Content(WorkspaceBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {

		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			resources.Query[name] = true
		case modconfig.BlockTypeControl:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			resources.Control[name] = true
		case modconfig.BlockTypeBenchmark:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			resources.Benchmark[name] = true
		}
	}
	return resources, nil
}
