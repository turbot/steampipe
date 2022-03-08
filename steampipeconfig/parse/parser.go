package parse

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
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
	for configPath, data := range fileData {
		var file *hcl.File
		var moreDiags hcl.Diagnostics
		ext := filepath.Ext(configPath)
		if ext == constants.JsonExtension {
			file, moreDiags = json.ParseFile(configPath)
		} else if constants.IsYamlExtension(ext) {
			file, moreDiags = parseYamlFile(configPath)
		} else {
			file, moreDiags = parser.ParseHCL(data, configPath)
		}

		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
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
func ParseModDefinition(modPath string) (*modconfig.Mod, error) {
	// TODO think about variables

	// if there is no mod at this location, return error
	modFilePath := filepaths.ModFilePath(modPath)
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no mod file found in %s", modPath)
	}
	fileData, diags := LoadFileData(modFilePath)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load mod files", diags)
	}

	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, diags := body.Content(ModBlockSchema)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	// build an eval context containing functions
	evalCtx := &hcl.EvalContext{
		Functions: ContextFunctions(modPath),
	}

	for _, block := range content.Blocks {
		if block.Type == modconfig.BlockTypeMod {
			var defRange = block.DefRange
			if hclBody, ok := block.Body.(*hclsyntax.Body); ok {
				defRange = hclBody.SrcRange
			}
			mod, err := modconfig.NewMod(block.Labels[0], modPath, defRange)
			if err != nil {
				return nil, err
			}
			diags = gohcl.DecodeBody(block.Body, evalCtx, mod)
			if diags.HasErrors() {
				return nil, plugin.DiagsToError("Failed to decode mod hcl file", diags)
			}
			// call decode callback
			if err := mod.OnDecoded(block, nil); err != nil {
				return nil, err
			}
			// set modfilename
			mod.SetFilePath(modFilePath)
			return mod, nil
		}
	}

	return nil, fmt.Errorf("no mod definition found in %s", modPath)
}

// ParseMod parses all source hcl files for the mod path and associated resources, and returns the mod object
// NOTE: the mod definition has already been parsed (or a default created) and is in opts.RunCtx.RootMod
func ParseMod(modPath string, fileData map[string][]byte, pseudoResources []modconfig.MappableResource, runCtx *RunContext) (*modconfig.Mod, error) {
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, moreDiags := body.Content(ModBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	mod := runCtx.CurrentMod
	if mod == nil {
		return nil, fmt.Errorf("ParseMod called with no Current Mod set in RunContext")
	}
	// get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources, err := loadMappableResourceNames(modPath, content)
	if err != nil {
		return nil, err
	}

	// if variables were passed in runcontext, add to the mod
	for _, v := range runCtx.Variables {
		if diags = mod.AddResource(v); diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to add resource to mod", diags)
		}
	}

	// add pseudo resources to the mod
	addPseudoResourcesToMod(pseudoResources, hclResources, mod)

	// add this mod to run context - this it to ensure all pseudo resources get added
	runCtx.SetDecodeContent(content, fileData)
	if diags = runCtx.AddMod(mod); diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to add mod to run context", diags)
	}

	// we may need to decode more than once as we gather dependencies as we go
	const maxDecodes = 3
	for attempt := 0; attempt < maxDecodes; attempt++ {
		if diags = decode(runCtx); diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to decode all mod hcl files", diags)
		}

		// if eval is complete, we're done
		if runCtx.EvalComplete() {
			break
		}
	}

	// we failed to resolve dependencies
	if !runCtx.EvalComplete() {
		str := runCtx.FormatDependencies()
		return nil, fmt.Errorf("failed to resolve mod dependencies\nDependencies:\n%s", str)
	}

	// now tell mod to build tree of controls.
	if err := mod.BuildResourceTree(runCtx.LoadedDependencyMods); err != nil {
		return nil, err
	}

	return mod, nil
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
func loadMappableResourceNames(modPath string, content *hcl.BodyContent) (map[string]bool, error) {
	hclResources := make(map[string]bool)

	// TODO update this to not have a single hardcoded pseudo resource type
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

	content, moreDiags := body.Content(ModBlockSchema)
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
