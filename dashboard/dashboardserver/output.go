package dashboardserver

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
)

var logSink io.Writer

const (
	errorPrefix   = "[ Error   ]"
	messagePrefix = "[ Message ]"
	readyPrefix   = "[ Ready   ]"
	waitPrefix    = "[ Wait    ]"
)

func initLogSink() {
	if viper.GetBool(constants.ArgServiceMode) {
		logName := fmt.Sprintf("dashboard-%s.log", time.Now().Format("2006-01-02"))
		logPath := filepath.Join(filepaths.EnsureLogDir(), logName)
		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("failed to open dashboard manager log file: %s\n", err.Error())
			os.Exit(3)
		}
		logSink = f
	} else {
		logSink = os.Stdout
	}
}

func output(_ context.Context, prefix string, msg interface{}) {
	fmt.Fprintf(logSink, "%s %v\n", prefix, msg)
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
