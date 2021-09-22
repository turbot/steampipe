package mod_installer

type ModInstaller struct {
	Workspace Workspace
}

func (i *ModInstaller) Install(mod string) {
	// parse mod into github url and tag name
	url, tag, error := i.ParseModName(mod)
}
