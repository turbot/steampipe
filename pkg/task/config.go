package task

import "context"

type TaskRunOption func(o *taskRunConfig)

type HookFn func(context.Context)

type taskRunConfig struct {
	preHooks       []HookFn
	runUpdateCheck bool
}

func newRunConfig() *taskRunConfig {
	return &taskRunConfig{
		runUpdateCheck: true,
	}
}

func WithUpdateCheck(run bool) TaskRunOption {
	return func(o *taskRunConfig) {
		o.runUpdateCheck = run
	}
}

func WithPreHook(f HookFn) TaskRunOption {
	return func(o *taskRunConfig) {
		o.preHooks = append(o.preHooks, f)
	}
}
