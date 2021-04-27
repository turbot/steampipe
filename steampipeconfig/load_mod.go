package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
)

// Op describes a set of file operations.
type LoadModFlag uint32

const (
	CreateDefaultMod LoadModFlag = 1 << iota
	CreatePseudoResources
)

type LoadModOptions struct {
	Flags       LoadModFlag
	ListOptions *filehelpers.ListOptions
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
func LoadMod(modPath string, opts *parse.ParseModOptions) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	if opts == nil {
		opts = &parse.ParseModOptions{}
	}
	// verify the mod folder exists
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mod folder %s does not exist", modPath)
	}

	var pseudoResources []modconfig.MappableResource
	// if flag is set, create pseudo resources by mapping files
	if opts.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		pseudoResources, err = createPseudoResources(modPath, opts)
		if err != nil {
			return nil, err
		}
	}

	// now parse the mod, passing the pseudo resources
	// load the raw data
	mod, err = parseMod(modPath, pseudoResources, opts)
	if err != nil {
		return nil, err
	}

	return
}

func parseMod(modPath string, pseudoResources []modconfig.MappableResource, opts *parse.ParseModOptions) (*modconfig.Mod, error) {
	// build list of all filepaths we need to parse/load
	// NOTE: pseudo resource creation is handled separately below
	opts.ListOptions.Include = filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension})
	sourcePaths, err := getSourcePaths(modPath, opts)
	if err != nil {
		log.Printf("[WARN] LoadMod: failed to get mod file paths: %v\n", err)
		return nil, err
	}

	fileData, diags := parse.LoadFileData(sourcePaths)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	// parse all hcl files.
	return parse.ParseModHcl(modPath, fileData, pseudoResources, opts)
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
func getSourcePaths(modPath string, opts *LoadModOptions) ([]string, error) {
	sourcePaths, err := filehelpers.ListFiles(modPath, opts.ListOptions)
	if err != nil {
		return nil, err
	}
	return sourcePaths, nil
}

// create pseudo-resources for any files whose extensions are registered
// NOTE: this mutates parseResults
func createPseudoResources(modPath string, opts *parse.ParseModOptions) ([]modconfig.MappableResource, error) {
	// list all registered files
	opts.ListOptions.Include = filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions())
	sourcePaths, err := getSourcePaths(modPath, opts)
	if err != nil {
		return nil, err
	}

	var errors []error
	var res []modconfig.MappableResource

	// for every source path:
	// - if it is NOT a registered type, skip
	// [- if an existing resource has already referred directly to this file, skip] *not yet*
	for _, path := range sourcePaths {
		factory, ok := modconfig.ResourceTypeMap[filepath.Ext(path)]
		if !ok {
			continue
		}
		resource, fileData, err := factory(modPath, path)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if resource != nil {
			metadata := getPseudoResourceMetadata(resource.Name(), path, fileData)
			resource.SetMetadata(metadata)
			res = append(res, resource)
		}
	}

	// show errors as trace logging
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[TRACE] failed to convert local file into resource: %v", err)
		}
	}

	return res, nil
}

func getPseudoResourceMetadata(name string, path string, fileData []byte) *modconfig.ResourceMetadata {
	sourceDefinition := string(fileData)
	split := strings.Split(sourceDefinition, "\n")
	lineCount := len(split)

	m := &modconfig.ResourceMetadata{
		ResourceName:     name,
		FileName:         path,
		StartLineNumber:  1,
		EndLineNumber:    lineCount,
		IsAutoGenerated:  true,
		SourceDefinition: sourceDefinition,
	}

	return m
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
	name := constants.WorkspaceDefaultModName
	return &modconfig.Mod{ShortName: &name}
}
