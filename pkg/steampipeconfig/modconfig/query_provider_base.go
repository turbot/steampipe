package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

type QueryProviderBase struct {
	HclResourceBase
	RuntimeDependencyProviderBase
	QueryProviderRemain hcl.Body `hcl:",remain" json:"-"`

	// TODO  [node_reuse] ONLY CONTROL HAS SQL AND QUERY JSON TAG
	// control
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `json:"-"`
	QueryName             *NamedItem  `cty:"query" hcl:"query" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	withs               []*DashboardWith
	runtimeDependencies map[string]*RuntimeDependency
	disableCtySerialise bool
}

// GetParams implements QueryProvider
func (b *QueryProviderBase) GetParams() []*ParamDef {
	return b.Params
}

// GetArgs implements QueryProvider
func (b *QueryProviderBase) GetArgs() *QueryArgs {
	return b.Args

}

// GetSQL implements QueryProvider
func (b *QueryProviderBase) GetSQL() *string {
	return b.SQL
}

// GetQuery implements QueryProvider
func (b *QueryProviderBase) GetQuery() *Query {
	return b.Query
}

// SetArgs implements QueryProvider
func (b *QueryProviderBase) SetArgs(args *QueryArgs) {
	b.Args = args
}

// SetParams implements QueryProvider
func (b *QueryProviderBase) SetParams(params []*ParamDef) {
	b.Params = params
}

// VerifyQuery implements QueryProvider
// returns an error if neither sql or query are set
// it is overidden by resource types for which sql is optional
func (b *QueryProviderBase) VerifyQuery(queryProvider QueryProvider) error {
	if queryProvider.GetQuery() == nil && queryProvider.GetSQL() == nil {
		return fmt.Errorf("%s must define either a 'sql' property or a 'query' property", queryProvider.Name())
	}
	return nil
}

// RequiresExecution implements QueryProvider
func (b *QueryProviderBase) RequiresExecution(queryProvider QueryProvider) bool {
	return queryProvider.GetQuery() != nil || queryProvider.GetSQL() != nil
}

// GetResolvedQuery return the SQL and args to run the query
func (b *QueryProviderBase) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
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

func (b *QueryProviderBase) SetQuery(q *Query) {
	b.Query = q
}

// MergeParentArgs merges our args with our parent args (ours take precedence)
func (b *QueryProviderBase) MergeParentArgs(queryProvider QueryProvider, parent QueryProvider) (diags hcl.Diagnostics) {
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
func (b *QueryProviderBase) GetQueryProviderBase() *QueryProviderBase {
	return b
}

// CtyValue implements CtyValueProvider
func (b *QueryProviderBase) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}
