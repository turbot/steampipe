package mod

import (
	"testing"

	"github.com/turbot/go-kit/helpers"
)

type resolutionTestCase struct {
	name     string
	mods     []*Mod
	expected *Resolution
}

var resolutionTestCases = map[string]*resolutionTestCase{

	"Single dependency, no versions": {
		mods: []*Mod{
			{
				Name:       "m1",
				ModDepends: []*Mod{{Name: "m2"}},
			},
			{
				Name: "m2",
			},
		},
		expected: &Resolution{
			Resolved: map[string]*Mod{
				"m1": {Name: "m1"},
				"m2": {Name: "m2"},
			},
		},
	},

	"Single dependency, versions": {
		mods: []*Mod{
			{
				Name:       "m1",
				Version:    "1.0",
				ModDepends: []*Mod{{Name: "m2", Version: "1.0"}},
			},
			{
				Name:    "m2",
				Version: "1.0",
			},
		},
		expected: &Resolution{
			Resolved: map[string]*Mod{
				"m1@1.0": {Name: "m1", Version: "1.0"},
				"m2@1.0": {Name: "m2", Version: "1.0"},
			},
		},
	},

	"Dependency with implicit latest version": {
		mods: []*Mod{
			{
				Name:       "m1",
				Version:    "1.0",
				ModDepends: []*Mod{{Name: "m2"}},
			},
			{
				Name:    "m2",
				Version: "1.0",
			},
		},
		expected: &Resolution{
			Resolved: map[string]*Mod{
				"m1@1.0": {Name: "m1", Version: "1.0"},
				"m2@1.0": {Name: "m2", Version: "1.0"},
			},
		},
	},

	"No dependencies": {
		mods: []*Mod{
			{
				Name: "m1",
			},
			{
				Name: "m2",
			},
		},
		expected: &Resolution{
			Resolved: map[string]*Mod{
				"m1": {Name: "m1"},
				"m2": {Name: "m2"},
			},
		},
	},
}

func TestResolveModDependencies(t *testing.T) {
	for name, test := range resolutionTestCases {

		resolution := ResolveModDependencies(test.mods)

		if !ResolutionsEqual(resolution, test.expected) {

			t.Errorf(`Test: '%s' FAILED : expected %v, got %v`, name, test.expected, resolution)
		}
	}
}

func ResolutionsEqual(l, r *Resolution) bool {

	for name, lResolved := range l.Resolved {
		if !ModsEqual(lResolved, r.Resolved[name]) {
			return false
		}
	}

	for name, lUnresolved := range l.Unresolved {
		if !UnresolvedModsEqual(lUnresolved, r.Unresolved[name]) {
			return false
		}
	}
	return true

}

func UnresolvedModsEqual(l, r *UnresolvedMod) bool {
	if l == nil || r == nil {
		return l == nil && r == nil
	}
	return ModsEqual(l.Mod, r.Mod)
}

func ModsEqual(l, r *Mod) bool {
	if l == nil || r == nil {
		return l == nil && r == nil
	}

	if l.FullName() != r.FullName() {
		return false
	}
	lDeps := []string{}
	for _, d := range l.ModDepends {
		lDeps = append(lDeps, d.FullName())
	}
	for _, d := range r.ModDepends {
		if !helpers.StringSliceContains(lDeps, d.FullName()) {
			return false
		}
	}
	return true
}
