package statushooks

import (
	"context"
	"fmt"

	"github.com/briandowns/spinner"
)

type StatusHook struct {
	spinner *spinner.Spinner
}

func WithStatusHook(ctx context.Context) context.Context { return ctx }
func GetStatusHook(ctx context.Context) *StatusHook      { return nil }

func NewSetStatus(ctx context.Context, status string) {}
func ShowMessage(ctx context.Context, message string) {
	hook := GetStatusHook(ctx)
	if hook != nil && hook.spinner.Active() {
		hook.spinner.Stop()
		defer hook.spinner.Start()
	}
	fmt.Println(message)
}

func ShowWarning(ctx context.Context, warning string) {}

func ShowSpinner(ctx context.Context) {
	hook := GetStatusHook(ctx)
	if hook == nil {
		return
	}
	hook.spinner.Start()
}

func HideSpinner(ctx context.Context) {
	hook := GetStatusHook(ctx)
	if hook == nil {
		return
	}
	hook.spinner.Stop()
}
