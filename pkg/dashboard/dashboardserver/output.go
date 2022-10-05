package dashboardserver

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
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
	if logSink == nil {
		logSink = os.Stdout
	}
	fmt.Fprintf(logSink, "%s %v\n", prefix, msg)
}

func OutputMessage(ctx context.Context, msg string) {
	output(ctx, applyColor(messagePrefix, color.HiGreenString), msg)
}

func OutputWarning(ctx context.Context, msg string) {
	output(ctx, applyColor(messagePrefix, color.RedString), msg)
}

func OutputError(ctx context.Context, err error) {
	output(ctx, applyColor(errorPrefix, color.RedString), err)
}

func outputReady(ctx context.Context, msg string) {
	output(ctx, applyColor(readyPrefix, color.GreenString), msg)
}

func OutputWait(ctx context.Context, msg string) {
	output(ctx, applyColor(waitPrefix, color.CyanString), msg)
}

func applyColor(str string, color func(format string, a ...interface{}) string) string {
	if !isatty.IsTerminal(os.Stdout.Fd()) || viper.GetBool(constants.ArgServiceMode) {
		return str
	} else {
		return color((str))
	}
}
