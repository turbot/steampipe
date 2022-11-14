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
	runtimeDependencies map[string]*RuntimeDependency

	// ONLY CONTROL HAS SQL AND QUERY JSON TAG
	// control
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`
	Mod                   *Mod        `cty:"mod" json:"-"`
	withs               []*DashboardWith
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

// GetMod implements QueryProvider
func (b *QueryProviderBase) GetMod() *Mod {
	return b.Mod
}

// GetPreparedStatementName implements QueryProvider
func (b *QueryProviderBase) GetPreparedStatementName() string {
	if b.PreparedStatementName != "" {
		return b.PreparedStatementName
	}
	b.PreparedStatementName = b.buildPreparedStatementName(b.ShortName, b.Mod.NameWithVersion(), constants.PreparedStatementImageSuffix)
	return b.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (b *QueryProviderBase) GetPreparedStatementExecuteSQL(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return b.getPreparedStatementExecuteSQL(b, runtimeArgs)
}

// VerifyQuery implements QueryProvider
// returns an error if neither sql or query are set
// it is overidden by resource types for which sql is optional
func (b *QueryProviderBase) VerifyQuery(queryProvider QueryProvider) error {
	// verify we have either SQL or a Query defined
	if queryProvider.GetQuery() == nil && queryProvider.GetSQL() == nil {
		// this should never happen as we should catch it in the parsing stage
		return fmt.Errorf("%s must define either a 'sql' property or a 'query' property", queryProvider.Name())
	}
	return nil
}

// RequiresExecution implements QueryProvider
func (b *QueryProviderBase) RequiresExecution(queryProvider QueryProvider) bool {
	return queryProvider.GetQuery() != nil || queryProvider.GetSQL() != nil
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

// return the SQLs to run the query as a prepared statement
func (b *QueryProviderBase) getResolvedQuery(queryProvider QueryProvider, runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	argsArray, err := ResolveArgs(queryProvider, runtimeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve args for %s: %s", queryProvider.Name(), err.Error())
	}
	sql := typehelpers.SafeString(queryProvider.GetSQL())
	// we expect there to be sql on the query provider, NOT a Query
	if sql == "" {
		return nil, fmt.Errorf("getResolvedQuery faiuled - no sql set for '%s'", queryProvider.Name())
	}

	return &ResolvedQuery{
		ExecuteSQL: sql,
		RawSQL:     sql,
		Args:       argsArray,
		Params:     queryProvider.GetParams(),
	}, nil
}

func (b *QueryProviderBase) AddRuntimeDependencies(dependencies []*RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		b.runtimeDependencies[dependency.String()] = dependency
	}
}

func (b *QueryProviderBase) MergeRuntimeDependencies(other QueryProvider) {
	dependencies := other.GetRuntimeDependencies()
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		if _, ok := b.runtimeDependencies[dependency.String()]; !ok {
			b.runtimeDependencies[dependency.String()] = dependency
		}
	}
}

func (b *QueryProviderBase) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
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

func (*QueryProviderBase) GetDescription() string {
	return ""
}

func (b *QueryProviderBase) AddWith(with *DashboardWith) {
	b.withs = append(b.withs, with)
}

func (b *QueryProviderBase) GetWith(name string) (*DashboardWith, bool) {
	for _, w := range b.withs {
		if w.UnqualifiedName == name {
			return w, true
		}
	}
	return nil, false
}
func (b *QueryProviderBase) GetWiths() []*DashboardWith {
	return b.withs
}
