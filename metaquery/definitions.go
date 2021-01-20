package metaquery

import "github.com/turbot/steampipe/constants"

type metaQueryDefinition struct {
	title       string
	description string
	args        []string
	handler     handler
	validator   validator
	completer   completer
}

var metaQueryDefinitions map[string]metaQueryDefinition

func init() {
	metaQueryDefinitions = map[string]metaQueryDefinition{
		cmdExit: {
			title:       cmdExit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		cmdQuit: {
			title:       cmdQuit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		cmdTableList: {
			title:       cmdTableList,
			handler:     listTables,
			validator:   atMostNArgs(1),
			description: "List or describe tables",
		},
		cmdSeparator: {
			title:       cmdSeparator,
			handler:     setViperConfigFromArg(constants.ArgSeparator),
			validator:   exactlyNArgs(1),
			description: "Set csv output separator",
		},
		cmdHeaders: {
			title:       "headers",
			handler:     setHeader,
			validator:   booleanValidator(cmdHeaders, allowedArgValues(false, constants.ValOn, constants.ValOff)),
			description: "Enable or disable column headers",
			args:        []string{constants.ValOn, constants.ValOff},
			completer:   booleanCompleter,
		},
		cmdMulti: {
			title:       "multi-line",
			handler:     setMultiLine,
			validator:   booleanValidator(cmdMulti, allowedArgValues(false, constants.ValOn, constants.ValOff)),
			description: "Enable or disable multiline mode",
			args:        []string{constants.ValOn, constants.ValOff},
			completer:   booleanCompleter,
		},
		cmdTiming: {
			title:       "timing",
			handler:     setTiming,
			validator:   booleanValidator(cmdTiming, allowedArgValues(false, constants.ValOn, constants.ValOff)),
			description: "Enable or disable query execution timing",
			args:        []string{constants.ValOn, constants.ValOff},
			completer:   booleanCompleter,
		},
		cmdOutput: {
			title:       cmdOutput,
			handler:     setViperConfigFromArg(constants.ArgOutput),
			validator:   composeValidator(exactlyNArgs(1), allowedArgValues(false, constants.ValJSON, constants.ValCSV, constants.ValTable)),
			description: "Set output format",
			args:        []string{constants.ValJSON, constants.ValCSV, constants.ValTable},
			completer:   outputCompleter,
		},
		cmdInspect: {
			title:       cmdInspect,
			handler:     inspect,
			validator:   atMostNArgs(1),
			description: "View connections, tables & column information",
			completer:   inspectCompleter,
		},
		cmdConnections: {
			title:       cmdConnections,
			handler:     listConnections,
			validator:   noArgs,
			description: "List active connections",
		},
		cmdClear: {
			title:       cmdClear,
			handler:     clearScreen,
			validator:   noArgs,
			description: "Clear the console",
		},
	}
}
