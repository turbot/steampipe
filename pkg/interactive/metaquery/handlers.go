package metaquery

import (
	"context"
	"fmt"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
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
	cmdconfig.Viper().Set(constants.ArgHeader, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// .multi
// set the ArgMulti viper key with the boolean value evaluated from arg[0]
func setMultiLine(_ context.Context, input *HandlerInput) error {
	cmdconfig.Viper().Set(constants.ArgMultiLine, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// .timing
// set the ArgHeader viper key with the boolean value evaluated from arg[0]
func setTiming(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		showTiming()
		return nil
	}

	switch input.args()[0] {
	case "on":
		cmdconfig.Viper().Set(constants.ArgTiming, true)
		cmdconfig.Viper().Set(constants.ArgVerboseTiming, false)
	case "off":
		cmdconfig.Viper().Set(constants.ArgTiming, false)
		cmdconfig.Viper().Set(constants.ArgVerboseTiming, false)
	case "verbose":
		cmdconfig.Viper().Set(constants.ArgTiming, true)
		cmdconfig.Viper().Set(constants.ArgVerboseTiming, true)
	}
	return nil
}

func showTiming() {
	timing := cmdconfig.Viper().GetBool(constants.ArgTiming)
	verboseTiming := cmdconfig.Viper().GetBool(constants.ArgVerboseTiming)
	timingString := "off"
	if timing {
		if verboseTiming {
			timingString = "verbose"
		} else {
			timingString = "on"

		}
	}
	fmt.Printf(
		`Timing is %s. Available options are: %s, %s, %s.`,
		constants.Bold(timingString),
		constants.Bold("on"),
		constants.Bold("off"),
		constants.Bold("verbose"),
	)
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
	cmdconfig.Viper().Set(constants.ArgAutoComplete, typeHelpers.StringToBool(input.args()[0]))
	return nil
}
