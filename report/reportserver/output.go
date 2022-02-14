package reportserver

import (
	"context"
	"fmt"

	"github.com/fatih/color"
)

const (
	errorPrefix = "[ Error   ]"
	//warningPrefix = "[ Warning ]"
	messagePrefix = "[ Message ]"
	readyPrefix   = "[ Ready   ]"
	waitPrefix    = "[ Wait    ]"
)

func output(_ context.Context, prefix string, msg interface{}) {
	fmt.Printf("%s %v\n", prefix, msg)
}

func outputMessage(ctx context.Context, msg string) {
	output(ctx, color.HiGreenString(messagePrefix), msg)
}

func outputError(ctx context.Context, err error) {
	output(ctx, color.RedString(errorPrefix), err)
}

func outputReady(ctx context.Context, msg string) {
	output(ctx, color.GreenString(readyPrefix), msg)
}

func outputWait(ctx context.Context, msg string) {
	output(ctx, color.CyanString(waitPrefix), msg)
}
