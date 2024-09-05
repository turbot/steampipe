package metaquery

import (
	"context"
	"fmt"
	constants2 "github.com/turbot/pipe-fittings/constants"
	"strings"

	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"golang.org/x/exp/maps"
)

type handler func(ctx context.Context, input *HandlerInput) error

// Handle handles a metaquery execution from the interactive client
func Handle(ctx context.Context, input *HandlerInput) error {
	cmd, _ := getCmdAndArgs(input.Query)
	metaQueryObj, found := metaQueryDefinitions[cmd]
	if !found {
		return fmt.Errorf("not sure how to handle '%s'", cmd)
	}
	handlerFunction := metaQueryObj.handler
	return handlerFunction(ctx, input)
}

// .header
// set the ArgHeader viper key with the boolean value evaluated from arg[0]
func setHeader(_ context.Context, input *HandlerInput) error {
	cmdconfig.Viper().Set(constants2.ArgHeader, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// .multi
// set the ArgMulti viper key with the boolean value evaluated from arg[0]
func setMultiLine(_ context.Context, input *HandlerInput) error {
	cmdconfig.Viper().Set(constants2.ArgMultiLine, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// .timing
// set the ArgHeader viper key with the boolean value evaluated from arg[0]
func setTiming(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		showTimingFlag()
		return nil
	}

	cmdconfig.Viper().Set(constants2.ArgTiming, input.args()[0])
	return nil
}

func showTimingFlag() {
	timing := cmdconfig.Viper().GetString(constants2.ArgTiming)

	fmt.Printf(`Timing is %s. Available options are: %s`,
		constants.Bold(timing),
		constants.Bold(strings.Join(maps.Keys(constants.QueryTimingValueLookup), ", ")))
	// add an empty line here so that the rendering buffer can start from the next line
	fmt.Println()

	return
}

// .separator and .output
// set the value of `viperKey` in `viper` with the value from `args[0]`
func setViperConfigFromArg(viperKey string) handler {
	return func(_ context.Context, input *HandlerInput) error {
		cmdconfig.Viper().Set(viperKey, input.args()[0])
		return nil
	}
}

// .exit
func doExit(_ context.Context, input *HandlerInput) error {
	input.ClosePrompt()
	return nil
}

// .clear
func clearScreen(_ context.Context, input *HandlerInput) error {
	input.Prompt.ClearScreen()
	return nil
}

// .autocomplete
func setAutoComplete(_ context.Context, input *HandlerInput) error {
	cmdconfig.Viper().Set(constants2.ArgAutoComplete, typeHelpers.StringToBool(input.args()[0]))
	return nil
}
