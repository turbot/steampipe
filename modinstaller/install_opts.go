package modinstaller

type InstallOpts struct {
	WorkspacePath string
	Command       string
	DryRun        bool
	ModArgs       []string
}
