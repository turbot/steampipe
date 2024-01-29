package inputvars

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig/var_config"
	"github.com/turbot/terraform-components/tfdiags"
)

// CollectVariableValues inspects the various places that configuration input variable
// values can come from and constructs a map ready to be passed to the
// backend as part of a Operation.
//
// This method returns diagnostics relating to the collection of the values,
// but the values themselves may produce additional diagnostics when finally
// parsed.
func CollectVariableValues(workspacePath string, variableFileArgs []string, variablesArgs []string, workspaceMod *modconfig.Mod) (map[string]UnparsedVariableValue, error) {
	workspaceModName := workspaceMod.ShortName
	var modNames = make(map[string]struct{})
	for _, m := range workspaceMod.ResourceMaps.Mods {
		modNames[m.ShortName] = struct{}{}
	}

	ret := map[string]UnparsedVariableValue{}

	// First we'll deal with environment variables
	// since they have the lowest precedence.
	// (apart from values in the mod Require proeprty, which are handled separately later)
	{
		env := os.Environ()
		for _, raw := range env {
			if !strings.HasPrefix(raw, constants.EnvInputVarPrefix) {
				continue
			}
			raw = raw[len(constants.EnvInputVarPrefix):] // trim the prefix

			eq := strings.Index(raw, "=")
			if eq == -1 {
				// Seems invalid, so we'll ignore it.
				continue
			}

			name := raw[:eq]
			rawVal := raw[eq+1:]

			ret[name] = unparsedVariableValueString{
				str:        rawVal,
				name:       name,
				sourceType: ValueFromEnvVar,
			}
		}
	}

	// Next up we have some implicit files that are loaded automatically
	// if they are present. There's the original terraform.tfvars
	// (constants.DefaultVarsFilename) along with the later-added search for all files
	// ending in .auto.spvars.
	defaultVarsPath := filepaths.DefaultVarsFilePath(workspacePath)
	if _, err := os.Stat(defaultVarsPath); err == nil {
		diags := addVarsFromFile(defaultVarsPath, ValueFromAutoFile, ret)
		if diags.HasErrors() {
			return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", defaultVarsPath), diags)
		}

	}

	if infos, err := os.ReadDir("."); err == nil {
		// "infos" is already sorted by name, so we just need to filter it here.
		for _, info := range infos {
			name := info.Name()
			if !isAutoVarFile(name) {
				continue
			}
			diags := addVarsFromFile(name, ValueFromAutoFile, ret)
			if diags.HasErrors() {
				return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", name), diags)
			}

		}
	}

	// Finally we process values given explicitly on the command line, either
	// as individual literal settings or as additional files to read.
	for _, fileArg := range variableFileArgs {
		diags := addVarsFromFile(fileArg, ValueFromNamedFile, ret)
		if diags.HasErrors() {
			return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to load variables from '%s'", fileArg), diags)
		}
	}

	var diags tfdiags.Diagnostics
	for _, variableArg := range variablesArgs {
		// Value should be in the form "name=value", where value is a
		// raw string whose interpretation will depend on the variable's
		// parsing mode.
		raw := variableArg
		eq := strings.Index(raw, "=")
		if eq == -1 {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				fmt.Sprintf("The given --var option %q is not correctly specified. It must be a variable name and value separated an equals sign: --var key=value", raw),
				"",
			))
			continue
		}

		name := raw[:eq]
		rawVal := raw[eq+1:]
		ret[name] = unparsedVariableValueString{
			str:        rawVal,
			name:       name,
			sourceType: ValueFromCLIArg,
		}
	}

	if diags.HasErrors() {
		return nil, error_helpers.DiagsToError(fmt.Sprintf("failed to evaluate var args:"), diags)
	}

	// check viper for any interactively added variables
	if varMap := viper.GetStringMap(constants.ConfigInteractiveVariables); varMap != nil {
		for name, rawVal := range varMap {
			// Value should be in the form "name=value", where value is a
			// raw string whose interpretation will depend on the variable's
			// parsing mode.
			ret[name] = UnparsedInteractiveVariableValue{
				Name:     name,
				RawValue: rawVal.(string),
			}
		}
	}

	// now map any variable names of form <modname>.<variablename> to <modname>.var.<varname>
	// - if any var value is qualified with the workspace mod, remove the qualification
	// - remove any variables which are not in the root mod or first level dependencies
	ret = transformVarNames(ret, workspaceModName, modNames)

	return ret, nil
}

// map any variable names of form <modname>.<variablename> to <modname>.var.<varname>
func transformVarNames(rawValues map[string]UnparsedVariableValue, workspaceModName string, modNames map[string]struct{}) map[string]UnparsedVariableValue {
	ret := make(map[string]UnparsedVariableValue, len(rawValues))
	for k, v := range rawValues {

		parts := strings.Split(k, ".")
		if len(parts) > 1 {
			if _, ok := modNames[parts[0]]; !ok {
				// NOTE: skip any variables which are not in the root mod or first level dependencies
				continue
			}
			if parts[0] == workspaceModName {
				k = parts[1]
			} else {
				k = fmt.Sprintf("%s.var.%s", parts[0], parts[1])
			}
		}

		ret[k] = v
	}
	return ret
}

func addVarsFromFile(filename string, sourceType ValueSourceType, to map[string]UnparsedVariableValue) tfdiags.Diagnostics {
	var diags tfdiags.Diagnostics

	src, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Failed to read variables file",
				fmt.Sprintf("Given variables file %s does not exist.", filename),
			))
		} else {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Failed to read variables file",
				fmt.Sprintf("Error while reading %s: %s.", filename, err),
			))
		}
		return diags
	}

	// replace syntax `<modname>.<varname>=<var_value>` with `___steampipe_<modname>_<varname>=<var_value>
	sanitisedSrc, depVarAliases := sanitiseVariableNames(src)

	var f *hcl.File
	var hclDiags hcl.Diagnostics

	// attempt to parse the config
	f, hclDiags = hclsyntax.ParseConfig(sanitisedSrc, filename, hcl.Pos{Line: 1, Column: 1})
	diags = diags.Append(hclDiags)
	if f == nil || f.Body == nil {
		return diags
	}

	// Before we do our real decode, we'll probe to see if there are any blocks
	// of type "variable" in this body, since it's a common mistake for new
	// users to put variable declarations in tfvars rather than variable value
	// definitions, and otherwise our error message for that case is not so
	// helpful.
	{
		content, _, _ := f.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "variable",
					LabelNames: []string{"name"},
				},
			},
		})
		for _, block := range content.Blocks {
			name := block.Labels[0]
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Variable declaration in .tfvars file",
				Detail:   fmt.Sprintf("A .tfvars file is used to assign values to variables that have already been declared in .tf files, not to declare new variables. To declare variable %q, place this block in one of your .tf files, such as variables.tf.\n\nTo set a value for this variable in %s, use the definition syntax instead:\n    %s = <value>", name, block.TypeRange.Filename, name),
				Subject:  &block.TypeRange,
			})
		}
		if diags.HasErrors() {
			// If we already found problems then JustAttributes below will find
			// the same problems with less-helpful messages, so we'll bail for
			// now to let the user focus on the immediate problem.
			return diags
		}
	}

	attrs, hclDiags := f.Body.JustAttributes()
	diags = diags.Append(hclDiags)

	for name, attr := range attrs {
		// check for aliases
		if alias, ok := depVarAliases[name]; ok {
			name = alias
		}
		to[name] = unparsedVariableValueExpression{
			expr:       attr.Expr,
			sourceType: sourceType,
		}
	}
	return diags
}

func sanitiseVariableNames(src []byte) ([]byte, map[string]string) {
	// replace syntax `<modname>.<varname>=<var_value>` with `____steampipe_mod_<modname>_<varname>____=<var_value>

	lines := strings.Split(string(src), "\n")
	// make map of varname aliases
	var depVarAliases = make(map[string]string)

	for i, line := range lines {

		r := regexp.MustCompile(`^ ?(([a-z0-9\-_]+)\.([a-z0-9\-_]+)) ?=`)
		captureGroups := r.FindStringSubmatch(line)
		if captureGroups != nil && len(captureGroups) == 4 {
			fullVarName := captureGroups[1]
			mod := captureGroups[2]
			varName := captureGroups[3]

			aliasedName := fmt.Sprintf("____steampipe_mod_%s_variable_%s____", mod, varName)
			depVarAliases[aliasedName] = fullVarName
			lines[i] = strings.Replace(line, fullVarName, aliasedName, 1)

		}
	}

	// now try again
	src = []byte(strings.Join(lines, "\n"))
	return src, depVarAliases
}

// unparsedVariableValueLiteral is a UnparsedVariableValue
// implementation that was actually already parsed (!). This is
// intended to deal with expressions inside "tfvars" files.
type unparsedVariableValueExpression struct {
	expr       hcl.Expression
	sourceType ValueSourceType
}

func (v unparsedVariableValueExpression) ParseVariableValue(mode var_config.VariableParsingMode) (*InputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	val, hclDiags := v.expr.Value(nil) // nil because no function calls or variable references are allowed here
	diags = diags.Append(hclDiags)

	rng := tfdiags.SourceRangeFromHCL(v.expr.Range())

	return &InputValue{
		Value:       val,
		SourceType:  v.sourceType,
		SourceRange: rng,
	}, diags
}

// unparsedVariableValueString is a UnparsedVariableValue
// implementation that parses its value from a string. This can be used
// to deal with values given directly on the command line and via environment
// variables.
type unparsedVariableValueString struct {
	str        string
	name       string
	sourceType ValueSourceType
}

func (v unparsedVariableValueString) ParseVariableValue(mode var_config.VariableParsingMode) (*InputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	val, hclDiags := mode.Parse(v.name, v.str)
	diags = diags.Append(hclDiags)

	return &InputValue{
		Value:      val,
		SourceType: v.sourceType,
	}, diags
}

// isAutoVarFile determines if the file ends with .auto.spvars or .auto.spvars.json
func isAutoVarFile(path string) bool {
	return strings.HasSuffix(path, constants.AutoVariablesExtension)
}
