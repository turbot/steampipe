package error_helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

func init() {
	color.Output = ansicolor.NewAnsiColorWriter(os.Stderr)
}

func WrapError(err error) error {
	if err == nil {
		return nil
	}
	return HandleCancelError(err)
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
	fmt.Fprintf(color.Error, "%s: %v\n", pconstants.ColoredErr, TransformErrorToSteampipe(err))
}

// ShowErrorWithMessage displays the given error nicely with the given message
func ShowErrorWithMessage(ctx context.Context, err error, message string) {
	if err == nil {
		return
	}
	err = HandleCancelError(err)
	statushooks.Done(ctx)
	fmt.Fprintf(color.Error, "%s: %s - %v\n", pconstants.ColoredErr, message, TransformErrorToSteampipe(err))
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
	return fmt.Errorf("%s", strings.TrimSpace(errString))
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
		err = fmt.Errorf("query timeout exceeded (%ds)", viper.GetInt(pconstants.ArgDatabaseQueryTimeout))
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
	fmt.Fprintf(color.Error, "%s: %v\n", pconstants.ColoredWarn, warning)
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

	combinedErrorString := map[string]struct{}{prefix: {}}
	for _, e := range errors {
		if e == nil {
			continue
		}
		combinedErrorString[e.Error()] = struct{}{}
	}

	return fmt.Errorf("%s", strings.Join(maps.Keys(combinedErrorString), "\n\t"))
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
	return fmt.Errorf("%s: %s\n", prefix, TransformErrorToSteampipe(err).Error())
}
