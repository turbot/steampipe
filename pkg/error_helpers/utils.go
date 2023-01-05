package error_helpers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/sperr"
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
		log.Printf("[ERROR] FailOnError: %+#v\n", sperr.Wrap(err))
		panic(err)
	}
}

func FailOnErrorWithMessage(err error, message string) {
	FailOnError(sperr.WrapWithMessage(err, message))
}

func ShowError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	err = HandleCancelError(err)
	statushooks.Done(ctx)
	log.Printf("[ERROR] Error: %+#v\n", sperr.Wrap(err))
	fmt.Fprintf(color.Output, "%s: %v\n", colorErr, sperr.Wrap(err))
}

// ShowErrorWithMessage displays the given error nicely with the given message
func ShowErrorWithMessage(ctx context.Context, err error, message string) {
	ShowError(ctx, sperr.WrapWithMessage(err, message))
}

// HandleCancelError modifies a context.Canceled error into a readable error that can
// be printed on the console
func HandleCancelError(err error) error {
	if IsCancelledError(err) {
		err = sperr.WrapWithMessage(err, "execution cancelled")
	}

	return err
}

func HandleQueryTimeoutError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		err = sperr.WrapWithMessage(err, "query timeout exceeded (%ds)", viper.GetInt(constants.ArgDatabaseQueryTimeout))
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
	if len(errors) == 0 {
		return nil
	}

	if allErrorsNil(errors...) {
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
		if e == nil {
			continue
		}
		combinedErrorString = append(combinedErrorString, e.Error())
	}
	return fmt.Errorf(strings.Join(combinedErrorString, "\n\t"))
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
	return CombineErrorsWithPrefix("", errors...)
}

func PrefixError(err error, prefix string) error {
	return sperr.WrapWithMessage(err, prefix)
}
