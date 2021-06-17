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
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, TransformErrorToSteampipe(err))
}

// ShowErrorWithMessage displays the given error nicely with the given message
func ShowErrorWithMessage(err error, message string) {
	err = handleCancelError(err)
	fmt.Fprintf(color.Output, "%s: %s - %v\n", colorErr, message, TransformErrorToSteampipe(err))
}

// TransformErrorToSteampipe removes the pq: and rpc error prefixes along
// with all the unnecessary information that comes from the
// drivers and libraries
func TransformErrorToSteampipe(err error) error {
	errString := strings.TrimSpace(err.Error())

	// an error that originated from our database/sql driver (always prefixed with "pq:")
	if strings.HasPrefix(errString, "pq:") {
		errString = strings.TrimSpace(strings.TrimPrefix(errString, "pq:"))

		// if this is an RPC Error while talking with the plugin
		if strings.HasPrefix(errString, "rpc error") {
			// trim out "rpc error: code = Unknown desc ="
			errString = strings.TrimPrefix(errString, "rpc error: code = Unknown desc =")
		}
	}
	return fmt.Errorf(strings.TrimSpace(errString))
}

// modifies a context.Canceled error into a readable error that can
// be printed on the console
func handleCancelError(err error) error {
	if err == context.Canceled {
		err = fmt.Errorf("execution cancelled")
	}
	return err
}

func ShowWarning(warning string) {
	fmt.Fprintf(color.Output, "%s: %v\n", colorWarn, warning)
}

func CombineErrorsWithPrefix(prefix string, errors ...error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return fmt.Errorf("%s%s", prefix, errors[0].Error())
	}

	combinedErrorString := []string{prefix}
	for _, e := range errors {
		combinedErrorString = append(combinedErrorString, e.Error())
	}
	return fmt.Errorf(strings.Join(combinedErrorString, "\n\t"))
}

func CombineErrors(errors ...error) error {
	return CombineErrorsWithPrefix("", errors...)
}
