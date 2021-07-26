package utils

import (
	"context"
	"errors"
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
		err = HandleCancelError(err)
		panic(err)
	}
}

func FailOnErrorWithMessage(err error, message string) {
	if err != nil {
		err = HandleCancelError(err)
		panic(fmt.Sprintf("%s: %s", message, err.Error()))
	}
}

func ShowError(err error) {
	err = HandleCancelError(err)
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, TransformErrorToSteampipe(err))
}

// ShowErrorWithMessage displays the given error nicely with the given message
func ShowErrorWithMessage(err error, message string) {
	err = HandleCancelError(err)
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

// HandleCancelError modifies a context.Canceled error into a readable error that can
// be printed on the console
func HandleCancelError(err error) error {
	if IsCancelledError(err) {
		err = fmt.Errorf("execution cancelled")
	}
	return err
}

func IsCancelledError(err error) bool {
	return errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "canceling statement due to user request")
}

func ShowWarning(warning string) {
	fmt.Fprintf(color.Output, "%s: %v\n", colorWarn, warning)
}

func CombineErrorsWithPrefix(prefix string, errors ...error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		if len(prefix) == 0 {
			return errors[0]
		} else {
			return fmt.Errorf("%s - %s", prefix, errors[0].Error())
		}
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

func PrefixError(err error, prefix string) error {
	return fmt.Errorf("%s: %s\n", prefix, TransformErrorToSteampipe(err).Error())
}
