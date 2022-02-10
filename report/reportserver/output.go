package reportserver

import (
	"context"
	"fmt"

	"github.com/fatih/color"
)

const (
	eRROR   = "[ Error   ]"
	wARNING = "[ Warning ]"
	mESSAGE = "[ Message ]"
	rEADY   = "[ Ready   ]"
	wAIT    = "[ Wait    ]"
)

var (
	outputErrorPrefix   = color.RedString(eRROR)
	outputWarningPrefix = color.YellowString(wARNING)
	outputMessagePrefix = color.HiGreenString(mESSAGE)
	outputReadyPrefix   = color.GreenString(rEADY)
	outputWaitPrefix    = color.CyanString(wAIT)
)

func output(_ context.Context, prefix string, msg interface{}) {
	fmt.Printf("%s %v\n", prefix, msg)
}

func outputMessage(ctx context.Context, msg string) {
	output(ctx, outputMessagePrefix, msg)
}

func outputError(ctx context.Context, err error) {
	output(ctx, outputErrorPrefix, err)
}

func outputReady(ctx context.Context, msg string) {
	output(ctx, outputReadyPrefix, msg)
}

func outputWait(ctx context.Context, msg string) {
	output(ctx, outputWaitPrefix, msg)
}
