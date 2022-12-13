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

	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `cty:"query" hcl:"query" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	// TACTICAL: store another reference to the base as a QueryProvider
	// stored purely so we can automatically determine whether we have overridden base properties
	baseQueryProvider QueryProvider

	withs               []*DashboardWith
	runtimeDependencies map[string]*RuntimeDependency
	disableCtySerialise bool
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

// GetQueryProviderBase implements QueryProvider
func (b *QueryProviderImpl) GetQueryProviderImpl() *QueryProviderImpl {
	return b
}

// ParamsInheritedFromBase implements QueryProvider
// determine whether our params were inherited from base resource
func (b *QueryProviderImpl) ParamsInheritedFromBase() bool {
	// note: this depends on baseQueryProvider being a reference to the same object as the derived class
	// base property which was used to populate the params
	if b.baseQueryProvider == nil {
		return false
	}

	baseParams := b.baseQueryProvider.GetParams()
	if len(b.Params) != len(baseParams) {
		return false
	}
	for i, p := range b.Params {
		if baseParams[i] != p {
			return false
		}
	}
	return true
}

// CtyValue implements CtyValueProvider
func (b *QueryProviderImpl) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}
