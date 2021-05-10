package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
)

var (
	colorErr    = color.RedString("Error")
	colorWarn   = color.YellowString("Warning")
	colorNotice = color.GreenString("Notice")
)

func init() {
	color.Output = ansicolor.NewAnsiColorWriter(os.Stderr)
}

func FailOnError(err error) {
	if err != nil {
		err = handleCancelError(err)
		panic(err)
	}
}
func FailOnErrorWithMessage(err error, message string) {
	if err != nil {
		err = handleCancelError(err)
		panic(fmt.Sprintf("%s: %s", message, err.Error()))
	}
}

func ShowError(err error) {
	err = handleCancelError(err)
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, TrimDriversFromErrMsg(err.Error()))
}

// ShowErrorWithMessage displays the given error nicely with the given message
func ShowErrorWithMessage(err error, message string) {
	err = handleCancelError(err)
	fmt.Fprintf(color.Output, "%s: %s - %v\n", colorErr, message, TrimDriversFromErrMsg(err.Error()))
}

// TrimDriversFromErrMsg removes the pq: and rpc error prefixes along
// with all the unnecessary information that comes from the
// drivers and libraries
func TrimDriversFromErrMsg(msg string) string {
	errString := strings.TrimSpace(msg)
	if strings.HasPrefix(errString, "pq:") {
		errString = strings.TrimSpace(strings.TrimPrefix(errString, "pq:"))
		if strings.HasPrefix(errString, "rpc error") {
			// trim out "rpc error: code = Unknown desc ="
			errString = strings.TrimSpace(errString[33:])
		}
	}
	return errString
}

func handleCancelError(err error) error {
	if err == context.Canceled {
		err = fmt.Errorf("query cancelled")
	}
	return err
}

func ShowWarning(warning string) {
	fmt.Fprintf(color.Output, "%s: %v\n", colorWarn, warning)
}
