package error_helpers

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
	sdk_error_helpers "github.com/turbot/steampipe-plugin-sdk/v5/error_helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/statushooks"
)

var (
	colorErr  = color.RedString("Error")
	colorWarn = color.YellowString("Warning")
)

func init() {
	color.Output = ansicolor.NewAnsiColorWriter(os.Stderr)
}

func WrapError(err error) error {
	if err == nil {
		return nil
	}
	return HandleCancelError(
		WrapPreparedStatementError(err))
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

func ShowError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	err = HandleCancelError(err)
	statushooks.Done(ctx)
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, TransformErrorToSteampipe(err))
}

// ShowErrorWithMessage displays the given error nicely with the given message
func ShowErrorWithMessage(ctx context.Context, err error, message string) {
	if err == nil {
		return
	}
	err = HandleCancelError(err)
	statushooks.Done(ctx)
	fmt.Fprintf(color.Output, "%s: %s - %v\n", colorErr, message, TransformErrorToSteampipe(err))
}

// TransformErrorToSteampipe removes the pq: and rpc error prefixes along
// with all the unnecessary information that comes from the
// drivers and libraries
func TransformErrorToSteampipe(err error) error {
	if err == nil {
		return err
	}
	// transform to a context
	err = HandleCancelError(err)

	errString := strings.TrimSpace(err.Error())

	// an error that originated from our database/sql driver (always prefixed with "ERROR:")
	if strings.HasPrefix(errString, "ERROR:") {
		errString = strings.TrimSpace(strings.TrimPrefix(errString, "ERROR:"))

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
		err = errors.New("execution cancelled")
	}

	return err
}

func HandleQueryTimeoutError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		err = fmt.Errorf("query timeout exceeded (%ds)", viper.GetInt(constants.ArgDatabaseQueryTimeout))
	}
	return err
}

func IsCancelledError(err error) bool {
	return errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "canceling statement due to user request")
}

func ShowWarning(warning string) {
	if len(warning) == 0 {
		return
	}
	fmt.Fprintf(color.Output, "%s: %v\n", colorWarn, warning)
}

func CombineErrorsWithPrefix(prefix string, errors ...error) error {
	return sdk_error_helpers.CombineErrorsWithPrefix(prefix, errors...)
}

func allErrorsNil(errors ...error) bool {
	for _, e := range errors {
		if e != nil {
			return false
		}
	}
	return true
}

func CombineErrors(errors ...error) error {
	return sdk_error_helpers.CombineErrors(errors...)
}

func PrefixError(err error, prefix string) error {
	return fmt.Errorf("%s: %s\n", prefix, TransformErrorToSteampipe(err).Error())
}
