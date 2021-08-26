package modconfig

import (
	"fmt"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
)

type ParamDef struct {
	Name        string  `hcl:"name,label"`
	Description *string `cty:"description" hcl:"description" column:"description,text"`
	Default     *string `cty:"default" hcl:"default" column:"default,text"`
}

func (d ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", d.Name, typehelpers.SafeString(d.Description), typehelpers.SafeString(d.Default))
}

func (q *Query) ResolveParams(params *QueryParams) (string, error) {
	var paramStrs, missingParams []string
	var err error
	if len(params.Params) > 0 {
		// do params contain named params?
		paramStrs, missingParams, err = q.resolveNamedParameters(params)
	} else {
		// resolve as positional parameters
		// (or fall back to defaults if no positional params are present)
		paramStrs, missingParams, err = q.resolvePositionalParameters(params)
	}

	if err != nil {
		return "", err
	}

	// did we resolve them all?
	if len(missingParams) > 0 {
		return "", fmt.Errorf("ResolveParams failed for %s - failed to resolve value for %d %s: %s",
			q.FullName,
			len(missingParams),
			utils.Pluralize("parameter", len(missingParams)),
			strings.Join(missingParams, ","))
	}

	// success!
	return fmt.Sprintf("(%s)", strings.Join(paramStrs, ",")), err
}

func (q *Query) resolveNamedParameters(params *QueryParams) (paramStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	if len(params.ParamsList) > 0 {
		err = fmt.Errorf("ResolveParams failed for %s - params data contain both positional and named parameters", q.FullName)
		return
	}
	// so params contain named params - if this query has no param defs, error out
	if len(q.ParamsDefs) < len(params.Params) {
		err = fmt.Errorf("ResolveParams failed for %s - params data contain %d named parameters but this query %d parameter definitions",
			q.FullName, len(params.Params), len(q.ParamsDefs))
		return
	}

	// to get here, we must have param defs for all provided named params
	paramStrs = make([]string, len(q.ParamsDefs))

	// iterate through each param def and resolve the value
	//for i, def := range q.ParamsDefs {
	//	defaultValue := typehelpers.SafeString(def.Default)
	//
	//	// can we resolve a value for this param?
	//	if val, ok := params.Params[def.Name]; ok {
	//		paramStrs[i] = val
	//	} else if defaultValue != "" {
	//		paramStrs[i] = defaultValue
	//	} else {
	//		// no value provided and no default defined - add to missing list
	//		missingParams = append(missingParams, def.Name)
	//	}
	//}

	return paramStrs, missingParams, nil
}

func (q *Query) resolvePositionalParameters(params *QueryParams) (paramStrs []string, missingParams []string, err error) {
	// if query params contains both positional and named params, error out
	if len(params.Params) > 0 {
		err = fmt.Errorf("ResolveParams failed for %s - params data contain both positional and named parameters", q.FullName)
		return
	}

	//// if no param defs are defined, just use the given values
	//if len(q.ParamsDefs) == 0 {
	//	paramStrs = params.ParamsList
	//	return
	//}
	//
	//// so there are param defs - we must be able to resolve all params
	//// if there are MORE defs than provided parameters, all remaining defs MUST provide a default
	//paramStrs = make([]string, len(q.ParamsDefs))
	//
	//for i, def := range q.ParamsDefs {
	//	defaultValue := typehelpers.SafeString(def.Default)
	//
	//	if i < len(params.ParamsList) {
	//		paramStrs[i] = params.ParamsList[i]
	//	} else if defaultValue != "" {
	//		// so we have run out of provided params - is there a default?
	//		paramStrs[i] = defaultValue
	//	} else {
	//		// no value provided and no default defined - add to missing list
	//		missingParams = append(missingParams, def.Name)
	//	}
	//}
	return
}
