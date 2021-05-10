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
		err = mutateErrorIfRequired(err)
		panic(err)
	}
}
func FailOnErrorWithMessage(err error, message string) {
	if err != nil {
		err = mutateErrorIfRequired(err)
		panic(fmt.Sprintf("%s: %s", message, err.Error()))
	}
}

func ShowError(err error) {
	err = mutateErrorIfRequired(err)
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, trimDriversFromErrMsg(err.Error()))
}

func ShowErrorWithMessage(err error, message string) {
	err = mutateErrorIfRequired(err)
	fmt.Fprintf(color.Output, "%s: %s - %v\n", colorErr, message, trimDriversFromErrMsg(err.Error()))
}

// remove the pq: and rpc error prefixes along
// with all the unnecessary information that comes from the
// drivers
func trimDriversFromErrMsg(msg string) string {
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
func mutateErrorIfRequired(err error) error {
	if err == context.Canceled {
		err = fmt.Errorf("query cancelled")
	}
	return err
}

func ShowWarning(warning string) {
	fmt.Fprintf(color.Output, "%s: %v\n", colorWarn, warning)
}
