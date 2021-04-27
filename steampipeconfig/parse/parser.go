package parse

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Parser struct {
	RootDirectory string
}

// built a map of filepath to file data
func LoadFileData(paths []string) (map[string][]byte, hcl.Diagnostics) {
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

// ParseModHcl :: parse all source hcl files for the mod and associated resources
func ParseModHcl(modPath string, fileData map[string][]byte, pseudoResources []modconfig.MappableResource, opts *ParseModOptions) (*modconfig.Mod, error) {

	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, moreDiags := body.Content(ModFileSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	// maps to store the parse resources
	var mod *modconfig.Mod

	// 1) get names of all resources defined in hcl
	hclResources := make(map[string]bool)
	for _, block := range content.Blocks {
		// if this is a mod, build a shell mod struct (with just thename populated)
		if block.Type == string(modconfig.BlockTypeMod) {
			// if there is more than one mod, fail
			if mod != nil {
				return nil, fmt.Errorf("more than 1 mod definition found in %s", modPath)
			}
			mod = modconfig.NewMod(block.Labels[0], modPath)
		} else {
			// all non mod resources, save the resource name
			// TODO ERROR HANDLING FOR BAD BLOCK TYPES
			name := modconfig.BuildModResourceName(modconfig.ModBlockType(block.Type), block.Labels[0])
			hclResources[name] = true
		}
	}

	// 2) create mod if needed
	if mod == nil {
		// should we create a default mod?
		if !opts.CreateDefaultMod() {
			// CreateDefaultMod flag NOT set - fail
			return nil, fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
		}
		// just create a default mod
		mod = defaultWorkspaceMod(modPath)
	}

	// 3) add pseudo resources to the mod
	var duplicates []string
	for _, r := range pseudoResources {
		// is there a hcl resource with the same name as this pseudo resource - it takes precedece
		if _, ok := hclResources[r.Name()]; ok {
			duplicates = append(duplicates, r.Name())
			continue
		}
		mod.AddPseudoResource(r)
	}
	if len(duplicates) > 0 {
		log.Printf("[TRACE] %d files were not converted into resources as hcl resources of same name are defined: %v", len(duplicates), duplicates)
	}

	// 4) Add dependencies?
	// TODO think about where we resolve and store mode dependencies

	// loop round the parsing until all dependencies have been resolved

	// create run context to handle dependency resolution
	runCtx := NewRunContext(mod, content.Blocks)

	for {
		blocks := runCtx.UnparsedBlocks
		runCtx.StartEvalLoop()

		// build evaluation context from mod and depends
		evalContext, err := runCtx.BuildEvalContext()
		if err != nil {
			return nil, err
		}

		for _, block := range blocks {
			blockType := modconfig.ModBlockType(block.Type)
			switch blockType {
			case modconfig.BlockTypeLocals:
				query, moreDiags := ParseQuery(block)
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					break
				}
				name := query.Name()
				if _, ok := mod.Queries[name]; ok {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("mod defines more that one query named %s", name),
						Subject:  &block.DefRange,
					})
					continue
				}
			case modconfig.BlockTypeMod:
				// pass the shell mod - it will be mutated
				modDepends, moreDiags := ParseMod(block, mod, evalContext)
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
				}
				if len(modDepends) > 0 {
					runCtx.AddUnparsedBlock(block)
					runCtx.AddDependencies(mod.Name(), modDepends)
				}

			case modconfig.BlockTypeQuery:
				query, moreDiags := ParseQuery(block)
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					break
				}
				name := query.Name()
				if _, ok := mod.Queries[name]; ok {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("mod defines more that one query named %s", name),
						Subject:  &block.DefRange,
					})
					continue
				}
				query.Metadata = GetMetadataForParsedResource(query.Name(), block, fileData, mod)
				mod.Queries[name] = query

			case modconfig.BlockTypeControl:
				control, moreDiags := ParseControl(block)
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					break
				}
				name := *control.ShortName
				if _, ok := mod.Controls[name]; ok {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("mod defines more that one control named %s", name),
						Subject:  &block.DefRange,
					})
					continue
				}
				control.Metadata = GetMetadataForParsedResource(control.Name(), block, fileData, mod)
				mod.Controls[name] = control

			case modconfig.BlockTypeControlGroup:
				controlGroup, moreDiags := ParseControlGroup(block)
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					break
				}
				name := types.SafeString(controlGroup.ShortName)
				if _, ok := mod.ControlGroups[name]; ok {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("mod defines more that one control group named %s", name),
						Subject:  &block.DefRange,
					})
					continue
				}
				controlGroup.Metadata = GetMetadataForParsedResource(controlGroup.Name(), block, fileData, mod)
				mod.ControlGroups[name] = controlGroup
			}
		}

		if diags.HasErrors() {
			return nil, plugin.DiagsToError("Failed to parse all mod hcl files", diags)
		}

		// continue looping until all blocks are parsed
		if runCtx.EvalComplete() {
			break
		}
	}

	// no tell mod to build tree of controls
	if err := mod.BuildControlTree(); err != nil {
		return nil, err
	}

	return mod, nil
}

func defaultWorkspaceMod(modPath string) *modconfig.Mod {
	return modconfig.NewMod("local", modPath)
}
