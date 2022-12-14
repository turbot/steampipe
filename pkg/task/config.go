package task

type taskRunConfig struct {
	runUpdateCheck bool
}

func newRunConfig() *taskRunConfig {
	return &taskRunConfig{
		runUpdateCheck: true,
	}
}

type TaskRunOption func(o *taskRunConfig)

func WithUpdateCheck(run bool) TaskRunOption {
	return func(o *taskRunConfig) {
		o.runUpdateCheck = run
	}
}
