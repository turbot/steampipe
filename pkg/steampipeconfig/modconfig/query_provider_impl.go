package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

type QueryProviderImpl struct {
	RuntimeDependencyProviderImpl
	QueryProviderRemain hcl.Body `hcl:",remain" json:"-"`

	SQL    *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query  *Query      `cty:"query" hcl:"query" json:"-"`
	Args   *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	withs               []*DashboardWith
	disableCtySerialise bool
	// flags to indicate if params and args were inherited from base resource
	baseArgs   bool
	baseParams bool
}

// GetParams implements QueryProvider
func (b *QueryProviderImpl) GetParams() []*ParamDef {
	return b.Params
}

// GetArgs implements QueryProvider
func (b *QueryProviderImpl) GetArgs() *QueryArgs {
	return b.Args

}

// GetSQL implements QueryProvider
func (b *QueryProviderImpl) GetSQL() *string {
	return b.SQL
}

// GetQuery implements QueryProvider
func (b *QueryProviderImpl) GetQuery() *Query {
	return b.Query
}

// SetArgs implements QueryProvider
func (b *QueryProviderImpl) SetArgs(args *QueryArgs) {
	b.Args = args
}

// SetParams implements QueryProvider
func (b *QueryProviderImpl) SetParams(params []*ParamDef) {
	b.Params = params
}

// ValidateQuery implements QueryProvider
// returns an error if neither sql or query are set
// it is overidden by resource types for which sql is optional
func (b *QueryProviderImpl) ValidateQuery() hcl.Diagnostics {
	var diags hcl.Diagnostics
	// Top level resources (with the exceptions of controls and queries) are never executed directly,
	// only used as base for a nested resource.
	// Therefore only nested resources, controls and queries MUST have sql or a query defined
	queryRequired := !b.IsTopLevel() ||
		helpers.StringSliceContains([]string{BlockTypeQuery, BlockTypeControl}, b.BlockType())

	if !queryRequired {
		return nil
	}

	if queryRequired && b.Query == nil && b.SQL == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s does not define a query or SQL", b.Name()),
			Subject:  b.GetDeclRange(),
		})
	}
	return diags
}

// RequiresExecution implements QueryProvider
func (b *QueryProviderImpl) RequiresExecution(queryProvider QueryProvider) bool {
	return queryProvider.GetQuery() != nil || queryProvider.GetSQL() != nil
}

// GetResolvedQuery return the SQL and args to run the query
func (b *QueryProviderImpl) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	argsArray, err := ResolveArgs(b, runtimeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve args for %s: %s", b.Name(), err.Error())
	}
	sql := typehelpers.SafeString(b.GetSQL())
	// we expect there to be sql on the query provider, NOT a Query
	if sql == "" {
		return nil, fmt.Errorf("getResolvedQuery faiuled - no sql set for '%s'", b.Name())
	}

	return &ResolvedQuery{
		ExecuteSQL: sql,
		RawSQL:     sql,
		Args:       argsArray,
	}, nil
}

// MergeParentArgs merges our args with our parent args (ours take precedence)
func (b *QueryProviderImpl) MergeParentArgs(queryProvider QueryProvider, parent QueryProvider) (diags hcl.Diagnostics) {
	parentArgs := parent.GetArgs()
	if parentArgs == nil {
		return nil
	}

	args, err := parentArgs.Merge(queryProvider.GetArgs(), parent)
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  parent.(HclResource).GetDeclRange(),
		}}
	}

	queryProvider.SetArgs(args)
	return nil
}

// GetQueryProviderImpl implements QueryProvider
func (b *QueryProviderImpl) GetQueryProviderImpl() *QueryProviderImpl {
	return b
}

// ParamsInheritedFromBase implements QueryProvider
// determine whether our params were inherited from base resource
func (b *QueryProviderImpl) ParamsInheritedFromBase() bool {
	return b.baseParams
}

// ArgsInheritedFromBase implements QueryProvider
// determine whether our args were inherited from base resource
func (b *QueryProviderImpl) ArgsInheritedFromBase() bool {
	return b.baseArgs
}

// CtyValue implements CtyValueProvider
func (b *QueryProviderImpl) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}

func (b *QueryProviderImpl) setBaseProperties() {
	b.RuntimeDependencyProviderImpl.setBaseProperties()
	if b.SQL == nil {
		b.SQL = b.getBaseImpl().SQL
	}
	if b.Query == nil {
		b.Query = b.getBaseImpl().Query
	}
	if b.Args == nil {
		b.Args = b.getBaseImpl().Args
		b.baseArgs = true
	}
	if b.Params == nil {
		b.Params = b.getBaseImpl().Params
		b.baseParams = true
	}
}

func (b *QueryProviderImpl) getBaseImpl() *QueryProviderImpl {
	return b.base.(QueryProvider).GetQueryProviderImpl()
}

func (b *QueryProviderImpl) MergeBaseDependencies(base QueryProvider) {
	//only merge dependency if target property of other was inherited
	//i.e. if other target propery
	baseRuntimeDependencies := base.GetRuntimeDependencies()
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, baseDep := range baseRuntimeDependencies {
		if _, ok := b.runtimeDependencies[baseDep.String()]; !ok {
			// was this target parent property (args/params) inherited
			if (baseDep.ParentPropertyName == "args" && !b.ArgsInheritedFromBase()) ||
				!b.ParamsInheritedFromBase() {
				continue
			}

			b.runtimeDependencies[baseDep.String()] = baseDep
		}
	}
}
