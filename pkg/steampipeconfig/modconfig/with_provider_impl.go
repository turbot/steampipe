package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"golang.org/x/exp/maps"
)

type WithProviderImpl struct {
	// required to allow partial decoding
	WithProviderRemain hcl.Body `hcl:",remain" json:"-"`

	// map of withs keyed by unqualified name
	withs map[string]*DashboardWith
}

func (b *WithProviderImpl) AddWith(with *DashboardWith) hcl.Diagnostics {
	if b.withs == nil {
		b.withs = make(map[string]*DashboardWith)
	}
	// if we already have this with, fail
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

func (b *WithProviderImpl) GetWiths() []*DashboardWith {
	return maps.Values(b.withs)
}

func (b *WithProviderImpl) GetWith(name string) (*DashboardWith, bool) {
	w, ok := b.withs[name]
	return w, ok
}
