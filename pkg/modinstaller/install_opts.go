package modinstaller

type InstallOpts struct {
	WorkspacePath string
	Command       string
	ModArgs       []string
	DryRun        bool
	Force         bool
}
