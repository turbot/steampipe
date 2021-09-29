package parse

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"sigs.k8s.io/yaml"
)

// LoadFileData builds a map of filepath to file data
func LoadFileData(paths ...string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var fileData = map[string][]byte{}

	for _, configPath := range paths {
		data, err := ioutil.ReadFile(configPath)

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

func ModfileExists(modPath string) bool {
	modFilePath := filepath.Join(modPath, "mod.sp")
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ParseModDefinition parses the modfile only
// it is expected the callign code wil lhave verified the existence of the modfile by calling ModfileExists
func ParseModDefinition(modPath string) (*modconfig.Mod, error) {
	// TODO think about variables

	// if there is no mod at this location, return error
	modFilePath := filepath.Join(modPath, "mod.sp")
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

	content, moreDiags := body.Content(ModBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	for _, block := range content.Blocks {
		if block.Type == modconfig.BlockTypeMod {
			mod := modconfig.NewMod(block.Labels[0], modPath, block.DefRange)
			diags := gohcl.DecodeBody(block.Body, nil, mod)
			if diags.HasErrors() {
				return nil, plugin.DiagsToError("Failed to decode mod hcl file", diags)
			}
			// call decode callback
			if err := mod.OnDecoded(block); err != nil {
				return nil, err
			}
			return mod, nil
		}
	}

	return nil, fmt.Errorf("no mod definition found in %s", modPath)
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

	src, err := ioutil.ReadAll(f)
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

// ParseMod parses all source hcl files for the mod path and associated resources, and returns the mod object
// NOTE: the mod definition has already been parsed (or a default created) and is in opts.RunCtx.RootMod
func ParseMod(modPath string, fileData map[string][]byte, pseudoResources []modconfig.MappableResource, opts *ParseModOptions) (*modconfig.Mod, error) {
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, moreDiags := body.Content(ModBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	mod := opts.RunCtx.CurrentMod
	if mod == nil {
		return nil, fmt.Errorf("ParseMod called with no Current Mod set in RunContext")
	}
	// get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources, err := loadMappableResourceNames(modPath, content)
	if err != nil {
		return nil, err
	}

	// add pseudo resources to the mod
	addPseudoResourcesToMod(pseudoResources, hclResources, mod)

	// add this mod to run context - this it to ensure all pseudo resources get added
	opts.RunCtx.AddMod(mod, content, fileData)

	// perform initial decode to get dependencies
	// (if there are no dependencies, this is all that is needed)
	diags = decode(opts)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to decode all mod hcl files", diags)
	}

	// if eval is not complete, there must be dependencies - run again in dependency order
	if !opts.RunCtx.EvalComplete() {
		diags = decode(opts)
		if diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to parse all mod hcl files", diags)
		}

		// we failed to resolve dependencies
		if !opts.RunCtx.EvalComplete() {
			return nil, fmt.Errorf("failed to resolve mod dependencies\nDependencies:\n%s", opts.RunCtx.FormatDependencies())
		}
	}

	// now tell mod to build tree of controls.
	// NOTE: this also builds the sorted benchmark list
	if err := mod.BuildResourceTree(); err != nil {
		return nil, err
	}

	return mod, nil
}

func addPseudoResourcesToMod(pseudoResources []modconfig.MappableResource, hclResources map[string]bool, mod *modconfig.Mod) {
	var duplicates []string
	for _, r := range pseudoResources {
		// is there a hcl resource with the same name as this pseudo resource - it takes precedence
		// TODO CHECK FOR PSEUDO RESOURCE DUPES AND WARN
		if _, ok := hclResources[r.Name()]; ok {
			duplicates = append(duplicates, r.Name())
			continue
		}
		// set mod pointer on pseudo resource
		r.SetMod(mod)
		// add pseudo resource to mod
		mod.AddPseudoResource(r)
	}
	if len(duplicates) > 0 {
		log.Printf("[TRACE] %d files were not converted into resources as hcl resources of same name are defined: %v", len(duplicates), duplicates)
	}
}

// get names of all resources defined in hcl which may also be created as pseudo resources
// if we find a mod block, build a shell mod
func loadMappableResourceNames(modPath string, content *hcl.BodyContent) (map[string]bool, error) {
	hclResources := make(map[string]bool)

	// TODO update thei to not have a single hardcoded pseudo resource type
	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {
		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			hclResources[name] = true
		}
		// TODO PANEL
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
			//case modconfig.BlockTypePanel:
			//	// for any mappable resource, store the resource name
			//	name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			//	resources.Panel[name]=true
		}
	}
	return resources, nil
}
