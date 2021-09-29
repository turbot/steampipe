package parse

import (
	goVersion "github.com/hashicorp/go-version"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
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
	RunCtx               *RunContext
	Flags                ParseModFlag
	ListOptions          *filehelpers.ListOptions
	LoadedDependencyMods modconfig.ModMap
	ModInstallationPath  string
	// if set, only decode these blocks
	BlockTypes []string
}

func NewParseModOptions(flags ParseModFlag, workspacePath string, listOptions *filehelpers.ListOptions) *ParseModOptions {
	return &ParseModOptions{
		Flags:                flags,
		ModInstallationPath:  constants.WorkspaceModPath(workspacePath),
		ListOptions:          listOptions,
		LoadedDependencyMods: make(modconfig.ModMap),
		RunCtx:               NewRunContext(workspacePath),
	}
}

func (o *ParseModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *ParseModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}
