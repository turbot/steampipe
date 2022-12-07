package task

type runConfig struct {
	runUpdateCheck bool
}

func newRunConfig() *runConfig {
	return &runConfig{
		runUpdateCheck: true,
	}
}

type TaskRunConfig func(o *runConfig)

func WithUpdateCheck(run bool) TaskRunConfig {
	return func(o *runConfig) {
		o.runUpdateCheck = run
	}
}
