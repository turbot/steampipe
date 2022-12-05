package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"golang.org/x/exp/maps"
)

type WithProviderBase struct {
	// map of withs keyed by unqualified name
	withs map[string]*DashboardWith
}

func (b *WithProviderBase) AddWith(with *DashboardWith) hcl.Diagnostics {
	if b.withs == nil {
		b.withs = make(map[string]*DashboardWith)
	}
	// if we already have a this with, fail
	if _, ok := b.withs[with.UnqualifiedName]; ok {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("duplicate with block '%s'", with.ShortName),
			Subject:  with.GetDeclRange(),
		}}
	}
	b.withs[with.UnqualifiedName] = with
	return nil
}

func (b *WithProviderBase) GetWiths() []*DashboardWith {
	return maps.Values(b.withs)
}
