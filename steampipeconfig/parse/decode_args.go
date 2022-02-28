package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func decodeArgs(attr *hcl.Attribute, evalCtx *hcl.EvalContext, resource modconfig.QueryProvider) (*modconfig.QueryArgs, []*modconfig.RuntimeDependency, hcl.Diagnostics) {
	var runtimeDependencies []*modconfig.RuntimeDependency
	var args = modconfig.NewQueryArgs()
	var diags hcl.Diagnostics

	v, valDiags := attr.Expr.Value(evalCtx)
	ty := v.Type()
	// determine which diags are runtime dependencies (which we allow) and which are not
	if valDiags.HasErrors() {
		for _, diag := range diags {
			dependency := diagsToDependency(diag)
			if dependency == nil || !dependency.IsRuntimeDependency() {
				diags = append(diags, diag)
			}
		}
	}
	// now diags contains all diags which are NOT runtime dependencies
	if diags.HasErrors() {
		return nil, nil, diags
	}

	var err error

	switch {
	case ty.IsObjectType():
		args.ArgMap, runtimeDependencies, err = ctyObjectToArgMap(attr, v, evalCtx)
	case ty.IsTupleType():
		args.ArgsList, runtimeDependencies, err = ctyTupleToArgArray(attr, v)
	default:
		err = fmt.Errorf("'params' property must be either a map or an array")
	}

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has invalid parameter config", resource.Name()),
			Detail:   err.Error(),
			Subject:  &attr.Range,
		})
	}
	return args, runtimeDependencies, diags
}

func ctyTupleToArgArray(attr *hcl.Attribute, val cty.Value) ([]*string, []*modconfig.RuntimeDependency, error) {
	// convert the attribute to a slice
	values := val.AsValueSlice()

	// build output array
	res := make([]*string, len(values))
	var runtimeDependencies []*modconfig.RuntimeDependency

	for idx, v := range values {
		// if the value is unknown, this is a runtime dependency
		if !v.IsKnown() {
			runtimeDependency, err := identifyRuntimeDependenciesFromArray(attr, idx)
			if err != nil {
				return nil, nil, err
			}

			runtimeDependencies = append(runtimeDependencies, runtimeDependency)
		} else {
			// decode the value into a postgres compatible
			valStr, err := ctyToPostgresString(v)
			if err != nil {
				err := fmt.Errorf("invalid value provided for arg #%d: %v", idx, err)
				return nil, nil, err
			}

			res[idx] = &valStr
		}
	}
	return res, runtimeDependencies, nil
}

func ctyObjectToArgMap(attr *hcl.Attribute, val cty.Value, evalCtx *hcl.EvalContext) (map[string]string, []*modconfig.RuntimeDependency, error) {
	res := make(map[string]string)
	var runtimeDependencies []*modconfig.RuntimeDependency
	it := val.ElementIterator()
	for it.Next() {
		k, v := it.Element()

		// decode key
		var key string
		if err := gocty.FromCtyValue(k, &key); err != nil {
			return nil, nil, err
		}

		// if the value is unknown, this is a runtime dependency
		if !v.IsKnown() {
			runtimeDependency, err := identifyRuntimeDependenciesFromObject(attr, key, evalCtx)
			if err != nil {
				return nil, nil, err
			}
			runtimeDependencies = append(runtimeDependencies, runtimeDependency)
		} else {
			// decode the value into a postgres compatible
			valStr, err := ctyToPostgresString(v)
			if err != nil {
				err := fmt.Errorf("invalid value provided for param '%s': %v", key, err)
				return nil, nil, err
			}

			res[key] = valStr
		}
	}
	return res, runtimeDependencies, nil
}

func identifyRuntimeDependenciesFromObject(attr *hcl.Attribute, key string, evalCtx *hcl.EvalContext) (*modconfig.RuntimeDependency, error) {
	// find the expression for this key
	argsExpr, ok := attr.Expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, fmt.Errorf("could not extract runtime dependency for arg %s", key)
	}
	for _, item := range argsExpr.Items {
		argNameCty, valDiags := item.KeyExpr.Value(evalCtx)
		if valDiags.HasErrors() {
			return nil, fmt.Errorf("could not extract runtime dependency for arg %s", key)
		}
		var argName string
		if err := gocty.FromCtyValue(argNameCty, &argName); err != nil {
			return nil, err
		}
		if argName == key {
			propertyPath, err := modconfig.ParseResourcePropertyPath(hclhelpers.TraversalAsString(item.ValueExpr.(*hclsyntax.ScopeTraversalExpr).Traversal))
			if err != nil {
				return nil, err
			}

			ret := &modconfig.RuntimeDependency{
				PropertyPath: propertyPath,
				ArgName:      &key,
			}
			return ret, nil
		}
	}
	return nil, fmt.Errorf("could not extract runtime dependency for arg %s - not found in attribute map", key)
}

func identifyRuntimeDependenciesFromArray(attr *hcl.Attribute, idx int) (*modconfig.RuntimeDependency, error) {
	// find the expression for this key
	argsExpr, ok := attr.Expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("could not extract runtime dependency for arg #%d", idx)
	}
	for i, item := range argsExpr.Exprs {
		if i == idx {
			propertyPath, err := modconfig.ParseResourcePropertyPath(hclhelpers.TraversalAsString(item.(*hclsyntax.ScopeTraversalExpr).Traversal))
			if err != nil {
				return nil, err
			}

			ret := &modconfig.RuntimeDependency{
				PropertyPath: propertyPath,
				ArgIndex:     &idx,
			}

			return ret, nil
		}
	}
	return nil, fmt.Errorf("could not extract runtime dependency for arg %d - not found in attribute list", idx)
}
