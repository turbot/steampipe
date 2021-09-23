package parse

import (
	goVersion "github.com/hashicorp/go-version"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
	CreatePseudoResources
)

type InstalledMod struct {
	Mod     *modconfig.Mod
	Version *goVersion.Version
}

type ParseModOptions struct {
	Flags                ParseModFlag
	ListOptions          *filehelpers.ListOptions
	LoadedDependencyMods modconfig.ModMap
	ModInstallationPath  string
	// if set, only decode these blocks
	BlockTypes []string
	RunCtx     *RunContext
	// the root mod which is being parsed
	// TODO only need it here as when we set it from LoadMod we do not have a runCtx yet
	//RootMod *modconfig.Mod
}

func NewParseModOptions(flags ParseModFlag, listOptions *filehelpers.ListOptions) *ParseModOptions {
	return &ParseModOptions{
		Flags:                flags,
		ListOptions:          listOptions,
		LoadedDependencyMods: make(modconfig.ModMap),
		RunCtx:               NewRunContext(),
	}
}

func (o *ParseModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *ParseModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}
