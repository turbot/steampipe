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
		constants.CmdHelp: {
			title:       constants.CmdHelp,
			handler:     doHelp,
			validator:   noArgs,
			description: "Show steampipe help",
		},
		constants.CmdExit: {
			title:       constants.CmdExit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		constants.CmdQuit: {
			title:       constants.CmdQuit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		constants.CmdTableList: {
			title:       constants.CmdTableList,
			handler:     listTables,
			validator:   atMostNArgs(1),
			description: "List or describe tables",
		},
		constants.CmdSeparator: {
			title:       constants.CmdSeparator,
			handler:     setViperConfigFromArg(constants.ArgSeparator),
			validator:   exactlyNArgs(1),
			description: "Set csv output separator",
		},
		constants.CmdHeaders: {
			title:       "headers",
			handler:     setHeader,
			validator:   booleanValidator(constants.CmdHeaders, validatorFromArgsOf(constants.CmdHeaders)),
			description: "Enable or disable column headers",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on headers in output"},
				{value: constants.ArgOff, description: "Turn off headers in output"},
			},
			completer: completerFromArgsOf(constants.CmdHeaders),
		},
		constants.CmdMulti: {
			title:       "multi-line",
			handler:     setMultiLine,
			validator:   booleanValidator(constants.CmdMulti, validatorFromArgsOf(constants.CmdMulti)),
			description: "Enable or disable multiline mode",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on multiline mode"},
				{value: constants.ArgOff, description: "Turn off multiline mode"},
			},
			completer: completerFromArgsOf(constants.CmdMulti),
		},
		constants.CmdTiming: {
			title:       "timing",
			handler:     setTiming,
			validator:   booleanValidator(constants.CmdTiming, validatorFromArgsOf(constants.CmdTiming)),
			description: "Enable or disable query execution timing",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Display time elapsed after every query"},
				{value: constants.ArgOff, description: "Turn off query timer"},
			},
			completer: completerFromArgsOf(constants.CmdTiming),
		},
		constants.CmdOutput: {
			title:       constants.CmdOutput,
			handler:     setViperConfigFromArg(constants.ArgOutput),
			validator:   composeValidator(exactlyNArgs(1), validatorFromArgsOf(constants.CmdOutput)),
			description: "Set output format: csv, json or table",
			args: []metaQueryArg{
				{value: constants.ArgJSON, description: "Set output to JSON"},
				{value: constants.ArgCSV, description: "Set output to CSV"},
				{value: constants.ArgTable, description: "Set output to Table"},
				{value: constants.ArgLine, description: "Set output to Line"},
			},
			completer: completerFromArgsOf(constants.CmdOutput),
		},
		constants.CmdCache: {
			title:       constants.CmdCache,
			handler:     cacheControl,
			validator:   composeValidator(exactlyNArgs(1), validatorFromArgsOf(constants.CmdCache)),
			description: "Toggle the query cache (on/off) or Clear the query cache",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on Caching"},
				{value: constants.ArgOff, description: "Turn off caching"},
				{value: constants.ArgClear, description: "Clear the cache"},
			},
			completer: completerFromArgsOf(constants.CmdCache),
		},
		constants.CmdInspect: {
			title:       constants.CmdInspect,
			handler:     inspect,
			validator:   atMostNArgs(1),
			description: "View connections, tables & column information",
			completer:   inspectCompleter,
		},
		constants.CmdConnections: {
			title:       constants.CmdConnections,
			handler:     listConnections,
			validator:   noArgs,
			description: "List active connections",
		},
		constants.CmdClear: {
			title:       constants.CmdClear,
			handler:     clearScreen,
			validator:   noArgs,
			description: "Clear the console",
		},
		constants.CmdSearchPath: {
			title:       constants.CmdSearchPath,
			handler:     setOrGetSearchPath,
			validator:   atMostNArgs(1),
			description: "Display the current search path, or set the search-path by passing in a comma-separated list",
		},
		constants.CmdSearchPathPrefix: {
			title:       constants.CmdSearchPathPrefix,
			handler:     setSearchPathPrefix,
			validator:   exactlyNArgs(1),
			description: "Set a prefix to the current search-path",
		},
	}
}
