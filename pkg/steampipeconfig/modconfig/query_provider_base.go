package modconfig

import (
	"fmt"
	"golang.org/x/exp/maps"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
)

type QueryProviderBase struct {
	runtimeDependencies map[string]*RuntimeDependency
	// map of withs keyed by unqualified name
	withs map[string]*DashboardWith
}

// VerifyQuery returns an error if neither sql or query are set
// it is overidden by resource types for which sql is optional
func (b *QueryProviderBase) VerifyQuery(queryProvider QueryProvider) error {
	// verify we have either SQL or a Query defined
	if queryProvider.GetQuery() == nil && queryProvider.GetSQL() == nil {
		// this should never happen as we should catch it in the parsing stage
		return fmt.Errorf("%s must define either a 'sql' property or a 'query' property", queryProvider.Name())
	}
	return nil
}

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

func (b *QueryProviderBase) AddWith(with *DashboardWith) hcl.Diagnostics {
	if b.withs == nil {
		b.withs = make(map[string]*DashboardWith)
	}
	if _, ok := b.GetWith(with.UnqualifiedName); ok {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("duplicate with block '%s'", with.ShortName),
			Subject:  with.GetDeclRange(),
		}}
	}
	b.withs[with.UnqualifiedName] = with
	return nil
}

func (b *QueryProviderBase) GetWith(name string) (*DashboardWith, bool) {
	w, ok := b.withs[name]
	return w, ok
}

func (b *QueryProviderBase) GetWiths() []*DashboardWith {
	return maps.Values(b.withs)
}
