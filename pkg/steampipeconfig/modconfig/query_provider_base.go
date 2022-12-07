package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
)

type QueryProviderBase struct {
	HclResourceBase
	RuntimeDependencyProviderBase
	QueryProviderRemain hcl.Body `hcl:",remain" json:"-"`

	// ONLY CONTROL HAS SQL AND QUERY JSON TAG
	// control
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	withs               []*DashboardWith
	runtimeDependencies map[string]*RuntimeDependency
	// we need the mod name for prepared statement name
	modNameWithVersion string
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

// GetPreparedStatementName implements QueryProvider
func (b *QueryProviderBase) GetPreparedStatementName() string {
	if b.PreparedStatementName != "" {
		return b.PreparedStatementName
	}
	b.PreparedStatementName = b.buildPreparedStatementName(b.ShortName, b.modNameWithVersion, constants.PreparedStatementImageSuffix)
	return b.PreparedStatementName
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

func (b *QueryProviderBase) GetQueryProviderBase() *QueryProviderBase {
	return b
}

func (b *QueryProviderBase) buildPreparedStatementName(queryName, modName, suffix string) string {
	// build prefix from mod name
	prefix := b.buildPreparedStatementPrefix(modName)

	// build the hash from the query/control name, mod name and suffix and take the first 4 bytes
	str := fmt.Sprintf("%s%s%s", prefix, queryName, suffix)
	hash := helpers.GetMD5Hash(str)[:4]
	// add hash to suffix
	suffix += hash

	// truncate the name if necessary
	nameLength := len(queryName)
	maxNameLength := constants.MaxPreparedStatementNameLength - (len(prefix) + len(suffix))
	if nameLength > maxNameLength {
		nameLength = maxNameLength
	}

	// construct the name
	return fmt.Sprintf("%s%s%s", prefix, queryName[:nameLength], suffix)
}

// set the prepared statement suffix and prefix
// and also store the parent resource object as a QueryProvider interface (base struct cannot cast itself to this)
func (b *QueryProviderBase) buildPreparedStatementPrefix(modName string) string {
	prefix := fmt.Sprintf("%s_", modName)
	prefix = strings.Replace(prefix, ".", "_", -1)
	prefix = strings.Replace(prefix, "@", "_", -1)

	return prefix
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
		// TODO KAI CHECK
		//Params: b.GetParams(),
	}, nil
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
