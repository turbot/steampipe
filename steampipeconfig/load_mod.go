package steampipeconfig

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

// parse all hcl files in modPath and return a single mod
// NOTE: it is an error if there is not exactly 1 mod resource in the folder
func LoadMod(modPath string) (mod *modconfig.Mod, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	// get all the config files in the directory
	sourcePaths, err := getFilePaths(modPath, constants.ModDataExtension)
	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get mod file paths: %v\n", err)
		return nil, err
	}
	if len(sourcePaths) == 0 {
		return nil, nil
	}
	fileData, diags := loadFileData(sourcePaths)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all mod files: %v\n", err)
		return nil, plugin.DiagsToError("Failed to load all mod files", diags)
	}

	body, diags := parseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all config files", diags)
	}

	content, moreDiags := body.Content(modFileSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load config", diags)
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
					Detail:   err.Error()})
				continue
			}
			queries[query.Name] = query
		}
	}

	// verify a mod has been parsed
	if mod == nil {
		return nil, fmt.Errorf("no mods found in %s", modPath)
	}

	if diags.HasErrors() {
		err = plugin.DiagsToError("Failed to load mod", diags)
	}
	mod.PopulateQueries(queries)
	return
}
