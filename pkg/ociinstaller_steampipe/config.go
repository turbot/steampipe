package ociinstaller_steampipe

type pluginInstallConfig struct {
	skipConfigFile bool
}

type pluginInstallOption = func(config *pluginInstallConfig)

func WithSkipConfig(skipConfigFile bool) pluginInstallOption {
	return func(o *pluginInstallConfig) {
		o.skipConfigFile = skipConfigFile
	}
}
