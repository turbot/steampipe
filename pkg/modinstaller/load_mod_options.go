package modinstaller

type loadModConfiguration struct {
	createDefault bool
}

// NewLoadModConfiguration creates a default configuration with createDefault
// set to false
func NewLoadModConfiguration() *loadModConfiguration {
	return &loadModConfiguration{
		createDefault: false,
	}
}

type LoadModOption = func(config *loadModConfiguration)

// WithCreateDefaultDisabled forcefully enables createDefault
func WithCreateDefault() LoadModOption {
	return func(config *loadModConfiguration) {
		config.createDefault = true
	}
}
