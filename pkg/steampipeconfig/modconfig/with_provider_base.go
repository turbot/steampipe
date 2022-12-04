package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"golang.org/x/exp/maps"
)

type WithProviderBase struct {
	// map of withs keyed by unqualified name
	withs  map[string]*DashboardWith
	parent ModTreeItem
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

//func (b *WithProviderBase) WalkParentPublishers(parentFunc func(WithProvider) (bool, error)) error {
//	for continueWalking := true; continueWalking; {
//		if parent := b.GetParentPublisher(); parent != nil {
//			var err error
//			continueWalking, err = parentFunc(parent)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (b *WithProviderBase) ResolveWithFromTree(name string) (*DashboardWith, bool) {
//
//	b.WalkParentPublishers(func(WithProvider) (bool, error)){
//
//	}
//	w, ok := b.withs[name]
//	if !ok {
//		parent := b.GetParentPublisher()
//		if parent != nil {
//			return parent.ResolveWithFromTree(name)
//		}
//	}
//	return w, ok
//}
//
//func (b *WithProviderBase) ResolveParamFromTree(name string) (any, bool) {
//	// TODO
//	return nil, false
//}

func (b *WithProviderBase) GetWiths() []*DashboardWith {
	return maps.Values(b.withs)
}

//func (b *WithProviderBase) GetParentPublisher() WithProvider {
//	parent := b.parent
//	for parent != nil {
//		if res, ok := parent.(WithProvider); ok {
//			return res
//		}
//		if grandparents := parent.GetParents(); len(grandparents) > 0 {
//			parent = grandparents[0]
//		}
//	}
//	return nil
//}
