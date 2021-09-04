package parse

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
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
		file, moreDiags := parser.ParseHCL(data, configPath)

		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
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

	// 1) get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources := make(map[string]bool)

	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just the name populated)
		switch block.Type {
		case modconfig.BlockTypeMod:
			// if there is more than one mod, fail
			if mod != nil {
				return nil, fmt.Errorf("more than 1 mod definition found in %s", modPath)
			}
			mod = modconfig.NewMod(block.Labels[0], modPath, block.DefRange)
		case modconfig.BlockTypeQuery:
			// for any mappable resource, store the resource name
			name := modconfig.BuildModResourceName(block.Type, block.Labels[0])
			hclResources[name] = true
		}
		// TODO PANEL
	}

	// 2) create mod if needed
	if mod == nil {
		// should we create a default mod?
		if !opts.CreateDefaultMod() {
			// CreateDefaultMod flag NOT set - fail
			return nil, fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
		}
		// just create a default mod
		mod = modconfig.CreateDefaultMod(modPath)
	}

	// 3) add pseudo resources to the mod
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

	// create run context to handle dependency resolution
	runCtx, diags := NewRunContext(mod, content, fileData, opts.Variables)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to create run context", diags)
	}

	// perform initial decode to get dependencies
	diags = decode(runCtx, opts)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to decode all mod hcl files", diags)
	}

	// if eval is not complete, there must be dependencies - run again in dependency order
	// (no need to do anything else here, this should be handled when building the eval context)
	if !runCtx.EvalComplete() {
		diags = decode(runCtx, opts)
		if diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to parse all mod hcl files", diags)
		}

		// we failed to resolve dependencies
		if !runCtx.EvalComplete() {
			return nil, fmt.Errorf("failed to resolve mod dependencies\nDependencies:\n%s", runCtx.FormatDependencies())
		}
	}

	// now tell mod to build tree of controls.
	// NOTE: this also builds the sorted benchmark list
	if err := mod.BuildResourceTree(); err != nil {
		return nil, err
	}

	return mod, nil
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
