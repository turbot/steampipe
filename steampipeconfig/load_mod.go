package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// Op describes a set of file operations.
type LoadModFlag uint32

const (
	CreateDefaultMod LoadModFlag = 1 << iota
	CreatePseudoResources
)

type LoadModOptions struct {
	Flags   LoadModFlag
	Exclude []string
}

func (o *LoadModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *LoadModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}

// LoadMod :: parse all hcl files in modPath and return a single mod
// if CreatePseudoResources flag is set, construct hcl resources for files with specific extensions
// NOTE: it is an error if there is more than 1 mod defined, however zero mods is acceptable
// - a default mod will be created assuming there are any resource files
func LoadMod(modPath string, opts *LoadModOptions) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	if opts == nil {
		opts = &LoadModOptions{}
	}
	// verify the mod folder exists
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", modPath)
	}

	// build list of all filepaths we need to parse/load
	// NOTE: pseudo resource creation is handled separately below
	var include = filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension})
	sourcePaths, err := getSourcePaths(modPath, include, opts.Exclude)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
		return nil, err
	}

	// load the raw data
	fileData, diags := loadFileData(sourcePaths)
	if diags.HasErrors() {
		log.Printf("[WARN] LoadMod: failed to load all mod files: %v\n", err)
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	// parse all hcl files.
	mod, err = parseModHcl(modPath, fileData, opts)
	if err != nil {
		return nil, err
	}

	// if flag is set, create pseudo resources by mapping files
	if opts.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		err = createPseudoResources(modPath, mod, opts)
		if err != nil {
			return nil, err
		}
	}

	return
}

// GetModFileExtensions :: return list of all file extensions we care about
// this will be the mod data extension, plus any registered extensions registered in fileToResourceMap
func GetModFileExtensions() []string {
	return append(modconfig.RegisteredFileExtensions(), constants.ModDataExtension)
}

// build list of all filepaths we need to parse/load the mod
// this will include hcl files (with .sp extension)
// as well as any other files with extensions that have been registered for pseudo resource creation
// (see steampipeconfig/modconfig/resource_type_map.go)
func getSourcePaths(modPath string, include, exclude []string) ([]string, error) {

	// build list options:
	// - search the current folder only
	// - include extensions we have identifed
	// - ignore the .steampipe folder
	opts := &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Exclude: exclude,
		Include: include,
	}
	sourcePaths, err := filehelpers.ListFiles(modPath, opts)
	if err != nil {
		return nil, err
	}
	return sourcePaths, nil
}

// parse all source hcl files for the mod and associated resources
func parseModHcl(modPath string, fileData map[string][]byte, opts *LoadModOptions) (*modconfig.Mod, error) {
	var mod *modconfig.Mod

	body, diags := parseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod source files", diags)
	}

	content, moreDiags := body.Content(modFileSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load mod", diags)
	}

	var queries = make(map[string]*modconfig.Query)
	var controls = make(map[string]*modconfig.Control)
	var controlGroups = make(map[string]*modconfig.ControlGroup)
	for _, block := range content.Blocks {
		blockType := modconfig.ModBlockType(block.Type)
		switch blockType {
		//case "variable":
		//	// TODO
		case modconfig.BlockTypeMod:
			// if there is more than one mod, fail
			if mod != nil {
				return nil, fmt.Errorf("more than 1 mod definition found in %s", modPath)
			}

			mod, moreDiags = parseMod(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}

		case modconfig.BlockTypeQuery:
			query, moreDiags := parseQuery(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			name := *query.ShortName
			if _, ok := queries[name]; ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("mod defines more that one query named %s", name),
					Subject:  &block.DefRange,
				})
				continue
			}
			queries[name] = query

		case modconfig.BlockTypeControl:
			control, moreDiags := parseControl(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			name := *control.ShortName
			if _, ok := controls[name]; ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("mod defines more that one control named %s", name),
					Subject:  &block.DefRange,
				})
				continue
			}
			controls[name] = control

		case modconfig.BlockTypeControlGroup:
			controlGroup, moreDiags := parseControlGroup(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			name := types.SafeString(controlGroup.ShortName)
			if _, ok := controlGroups[name]; ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("mod defines more that one control group named %s", name),
					Subject:  &block.DefRange,
				})
				continue
			}
			controlGroups[name] = controlGroup
		}
	}

	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to parse all mod hcl files", diags)
	}

	// is there a mod resource definition?
	if mod == nil {
		// should we creaste a default mod?
		if !opts.CreateDefaultMod() {
			// CreateDefaultMod flag NOT set - fail
			return nil, fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
		}
		// just create a default mod
		mod = defaultWorkspaceMod()
	}
	mod.Queries = queries
	mod.Controls = controls
	mod.ControlGroups = controlGroups

	// no tell mod to build tree of controls
	if err := mod.BuildControlTree(); err != nil {
		return nil, err
	}

	return mod, nil
}

// create pseudo-resources for any files whose extensions are registered
// NOTE: this mutates parseResults
func createPseudoResources(modPath string, mod *modconfig.Mod, opts *LoadModOptions) error {
	// list all registered files
	var include = filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions())
	sourcePaths, err := getSourcePaths(modPath, include, opts.Exclude)
	if err != nil {
		return err
	}

	var errors []error
	// for every source path:
	// - if it is NOT a registered type, skip
	// [- if an existing resource has already referred directly to this file, skip] *not yet*
	for _, path := range sourcePaths {
		factory, ok := modconfig.ResourceTypeMap[filepath.Ext(path)]
		if !ok {
			continue
		}
		resource, err := factory(modPath, path)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if resource != nil {
			if err := addResourceIfUnique(resource, mod, path); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// show errors as trace logging
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[TRACE] failed to convert local file into resource: %v", err)
		}
	}

	return nil
}

// add resource to parse results, if there is no resource of same name
func addResourceIfUnique(resource modconfig.MappableResource, mod *modconfig.Mod, path string) error {
	switch r := resource.(type) {
	case *modconfig.Query:
		// check there is not already a query with the same name
		if _, ok := mod.Queries[*r.ShortName]; ok {
			// we have already created a query with this name - skip!
			return fmt.Errorf("not creating resource for '%s' as there is already a query '%s' defined", path, *r.ShortName)
		}
		mod.Queries[*r.ShortName] = r
	}
	return nil
}

func defaultWorkspaceMod() *modconfig.Mod {
	defaultName := "local"
	return &modconfig.Mod{ShortName: &defaultName}
}
