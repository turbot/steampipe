package steampipeconfig

import (
	"fmt"
	"log"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

// parse all hcl files in modPath and return a single mod
// also construct hcl resources for files with specific extensions, as define by fileToResourceMap
// NOTE: it is an error if there is not exactly 1 mod resource in the folder
func LoadMod(modPath string) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	// build list of all filepaths we need to parse/load
	sourcePaths := getSourcePaths(modPath, err)
	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get mod file paths: %v\n", err)
		return nil, err
	}
	if len(sourcePaths) == 0 {
		return nil, nil
	}
	// load the raw data
	fileData, diags := loadFileData(sourcePaths)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all mod files: %v\n", err)
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	// parse all hcl files
	parseResult, err := parseModHcl(modPath, fileData)
	if err != nil {
		return nil, err
	}

	// if no mod is explicitly define in hcl, create a default
	if parseResult.mod != nil {
		mod = parseResult.mod
	} else {
		// create default workspace mod
		mod = defaultWorkspaceMod()
	}

	// now execute any psuedo-resource creations based on file mappings
	if err := createPseudoResources(parseResult, sourcePaths); err != nil {
		return nil, err
	}

	// now convert query map into an array and set on the mod object
	mod.PopulateQueries(parseResult.queryMap)
	return
}

type modParseResult struct {
	queryMap map[string]*modconfig.Query
	mod      *modconfig.Mod
}

// build list of all filepaths we need to parse/load
// this will include hcl files (with .sp extension) as well as any other files
// with extensions that have been registered for implicit resource creation
// (see steampipeconfig/modconfig/mappable_resource.go)
func getSourcePaths(modPath string, err error) []string {
	// build list of file extensions we care about
	// this will be the mod data extension, plus any registered extensions registered in fileToResourceMap
	var extensions = append(modconfig.RegisteredFileExtensions(), constants.ModDataExtension)

	// build include string from extensions
	var includeStrings []string
	for _, extension := range extensions {
		includeStrings = append(includeStrings, fmt.Sprintf("**/*%s", extension))
	}

	// build list options:
	// - searchg rescursivlet
	// - include extensions we have identifed
	// - ignore the .steampipe folder
	opts := &filehelpers.ListFilesOptions{
		Options: filehelpers.FilesRecursive,
		Exclude: []string{fmt.Sprintf("**/%s*", constants.WorkspaceDataDir)},
		Include: includeStrings,
	}
	sourcePaths, err := filehelpers.ListFiles(modPath, opts)
	return sourcePaths
}

// parse all source hck files for th emod and associated resources
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
func createPseudoResources(parseResults *modParseResult, sourcePaths []string) error {
	// for every source path:
	// - if it is NOT a registered type, skip
	// [- if an existing resource has already referred directly to this file, skip] *not yet*
	for _, path := range sourcePaths {
		factory, ok := modconfig.ResourceTypeMap[filepath.Ext(path)]
		if !ok {
			continue
		}
		resource, err := factory(path)
		if err != nil {
			return err
		}
		addResourceIfUnique(resource, parseResults)
	}
	return nil
}

func addResourceIfUnique(resource modconfig.MappableResource, parseResults *modParseResult) {
	switch r := resource.(type) {
	case *modconfig.Query:
		// check there is not already a query with the same name
		if _, ok := parseResults.queryMap[r.Name]; ok {
			// we have already created a query with this name - skip!
			return
		}
		parseResults.queryMap[r.Name] = r
	}
}

func defaultWorkspaceMod() *modconfig.Mod {
	return &modconfig.Mod{Name: "local"}
}
