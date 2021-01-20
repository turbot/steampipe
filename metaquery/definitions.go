package metaquery

import "github.com/turbot/steampipe/constants"

type metaQueryArg struct {
	value       string
	description string
}

type metaQueryDefinition struct {
	title       string
	description string
	args        []metaQueryArg
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
			validator:   booleanValidator(cmdHeaders, validatorFromArgsOf(cmdHeaders)),
			description: "Enable or disable column headers",
			args: []metaQueryArg{
				metaQueryArg{value: constants.ValOn, description: "Turn on headers in output"},
				metaQueryArg{value: constants.ValOff, description: "Turn off headers in output"},
			},
			completer: completerFromArgsOf(cmdHeaders),
		},
		cmdMulti: {
			title:       "multi-line",
			handler:     setMultiLine,
			validator:   booleanValidator(cmdMulti, validatorFromArgsOf(cmdMulti)),
			description: "Enable or disable multiline mode",
			args: []metaQueryArg{
				metaQueryArg{value: constants.ValOn, description: "Turn on multiline mode"},
				metaQueryArg{value: constants.ValOff, description: "Turn off multiline mode"},
			},
			completer: completerFromArgsOf(cmdMulti),
		},
		cmdTiming: {
			title:       "timing",
			handler:     setTiming,
			validator:   booleanValidator(cmdTiming, validatorFromArgsOf(cmdTiming)),
			description: "Enable or disable query execution timing",
			args: []metaQueryArg{
				metaQueryArg{value: constants.ValOn, description: "Display time elapsed after every query"},
				metaQueryArg{value: constants.ValOff, description: "Turn off query timer"},
			},
			completer: completerFromArgsOf(cmdTiming),
		},
		cmdOutput: {
			title:       cmdOutput,
			handler:     setViperConfigFromArg(constants.ArgOutput),
			validator:   composeValidator(exactlyNArgs(1), validatorFromArgsOf(cmdOutput)),
			description: "Set output format",
			args: []metaQueryArg{
				metaQueryArg{value: constants.ValJSON, description: "Set output to JSON"},
				metaQueryArg{value: constants.ValCSV, description: "Set output to CSV"},
				metaQueryArg{value: constants.ValTable, description: "Set output to Table"},
			},
			completer: completerFromArgsOf(cmdOutput),
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
