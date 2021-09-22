package parse

import (
	goVersion "github.com/hashicorp/go-version"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
	CreatePseudoResources
	LoadDependencies
)

type InstalledMod struct {
	Mod     *modconfig.Mod
	Version *goVersion.Version
}

type ParseModOptions struct {
	Flags               ParseModFlag
	ListOptions         *filehelpers.ListOptions
	Variables           map[string]cty.Value
	LoadedMods          modconfig.ModMap
	ModInstallationPath string
	// if set, only decode these blocks
	BlockTypes []string
	RunCtx     *RunContext
	// the root mod which is being parsed
	// TODO only need it here as when we set it from LoadMod we do not have a runCtx yet
	//RootMod *modconfig.Mod
}

func NewParseModOptions() *ParseModOptions {
	return &ParseModOptions{
		LoadedMods: make(modconfig.ModMap),
	}
}

func (o *ParseModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *ParseModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}

func (o *ParseModOptions) LoadDependencies() bool {
	return o.Flags&LoadDependencies == LoadDependencies
}
