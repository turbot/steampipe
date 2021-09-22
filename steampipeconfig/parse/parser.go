package parse

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
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

	// maps to store the parse resources
	var mod *modconfig.Mod

	// get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources, mod, err := loadShellModAndMappableResourceNames(modPath, content, mod)
	if err != nil {
		return nil, err
	}

	// create mod if needed
	if mod == nil {
		// should we create a default mod?
		if !opts.CreateDefaultMod() {
			// CreateDefaultMod flag NOT set - fail
			return nil, fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
		}
		// just create a default mod
		mod = modconfig.CreateDefaultMod(modPath)
	}

	// add pseudo resources to the mod
	addPseudoResourcesToMod(pseudoResources, hclResources, mod)

	// TODO not lovely that we set this in 2 places - here and LoadMod
	// if we have not set the root mod, this must be the root mod
	// (this is only needed if we are loading the workspace definition or variables
	// - if this is being called from LoadMod the RootMod will already be set)
	if opts.RootMod == nil {
		opts.RootMod = mod
	}
	// if we do not already have one, create run context to handle dependency resolution
	if opts.RunCtx == nil {
		runCtx, diags := NewRunContext(opts.RootMod, content, fileData, opts.Variables)
		if diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to create run context", diags)
		}
		opts.RunCtx = runCtx
	}

	// set the current mod
	opts.RunCtx.CurrentMod = mod
	// add this mod to run context
	opts.RunCtx.AddMod(mod)

	// perform initial decode to get dependencies
	// (if there are no depdnencies, this is all that is needed)
	diags = decode(opts)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to decode all mod hcl files", diags)
	}

	// if eval is not complete, there must be dependencies - run again in dependency order
	// (no need to do anything else here, this should be handled when building the eval context)
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
func loadShellModAndMappableResourceNames(modPath string, content *hcl.BodyContent, mod *modconfig.Mod) (map[string]bool, *modconfig.Mod, error) {
	hclResources := make(map[string]bool)

	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {
		case modconfig.BlockTypeMod:
			// if there is more than one mod, fail
			if mod != nil {
				return nil, nil, fmt.Errorf("more than 1 mod definition found in %s", modPath)
			}
			mod = modconfig.NewMod(block.Labels[0], modPath, block.DefRange)
		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			hclResources[name] = true
		}
		// TODO PANEL
	}
	return hclResources, mod, nil
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
