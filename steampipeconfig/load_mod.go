package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
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

// parse all hcl files in modPath and return a single mod
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

	// parse all hcl files
	parseResult, err := parseModHcl(modPath, fileData)
	if err != nil {
		return nil, err
	}

	// is there a mod resource definition?
	if parseResult.mod != nil {
		mod = parseResult.mod
	} else {
		// so there is no mod resource definition - check flag for our response
		if !opts.CreateDefaultMod() {
			// CreateDefaultMod flag NOT set - fail
			return nil, fmt.Errorf("mod folder %s does not contain a mod resource definition", modPath)
		}
		// just create a default mod
		mod = defaultWorkspaceMod()
	}

	// if flag is set, create pseudo resources by mapping files
	if opts.CreatePseudoResources() {
		// now execute any pseudo-resource creations based on file mappings
		err = createPseudoResources(modPath, parseResult, opts)
		if err != nil {
			return nil, err
		}
	}

	// now convert query map into an array and set on the mod object
	mod.PopulateQueries(parseResult.queryMap)
	return
}

type modParseResult struct {
	queryMap map[string]*modconfig.Query
	mod      *modconfig.Mod
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
	// - search recursively
	// - include extensions we have identifed
	// - ignore the .steampipe folder
	opts := &filehelpers.ListFilesOptions{
		Options: filehelpers.FilesRecursive,
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
func parseModHcl(modPath string, fileData map[string][]byte) (*modParseResult, error) {
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
	for _, block := range content.Blocks {
		switch block.Type {
		case "variable":
			// TODO
		case "mod":
			// if there is more than one mod, fail
			if mod != nil {
				return nil, fmt.Errorf("more than 1 mod definition found in %s", modPath)
			}

			mod, moreDiags = parseMod(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		case "query":
			query, moreDiags := parseQuery(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			if _, ok := queries[query.Name]; ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("mod defines more that one query named %s", query.Name),
					Subject:  &block.DefRange,
				})
				continue
			}
			queries[query.Name] = query
		}
	}

	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to parse all mod hcl files", diags)
	}

	return &modParseResult{queries, mod}, nil
}

// create pseudo-resources for any files whose extensions are registered
// NOTE: this mutates parseResults
func createPseudoResources(modPath string, parseResults *modParseResult, opts *LoadModOptions) error {

	// list all registered files
	var include = filehelpers.InclusionsFromExtensions(modconfig.RegisteredFileExtensions())
	sourcePaths, err := getSourcePaths(modPath, include, opts.Exclude)
	if err != nil {
		return err
	}

	// TODO currently we just add in unique results and ignore non-unique results
	// TODO ADD WARNING

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
		}
		addResourceIfUnique(resource, parseResults)
	}

	// TODO handle errors - show warning??
	return nil
}

// add resource to parse results, if there is no resource of same name
func addResourceIfUnique(resource modconfig.MappableResource, parseResults *modParseResult) bool {
	switch r := resource.(type) {
	case *modconfig.Query:
		// check there is not already a query with the same name
		if _, ok := parseResults.queryMap[r.Name]; ok {
			// we have already created a query with this name - skip!
			return false
		}
		parseResults.queryMap[r.Name] = r
	}
	return true
}

func defaultWorkspaceMod() *modconfig.Mod {
	return &modconfig.Mod{Name: "local"}
}
