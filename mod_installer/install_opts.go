package mod_installer

type InstallOpts struct {
	WorkspacePath string
	Command       string
	DryRun        bool
	ModArgs       []string
}
