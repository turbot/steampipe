package metaquery

import (
	"github.com/turbot/steampipe/pkg/constants"
)

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
		constants_steampipe.CmdHelp: {
			title:       constants_steampipe.CmdHelp,
			handler:     doHelp,
			validator:   noArgs,
			description: "Show steampipe help",
		},
		constants_steampipe.CmdExit: {
			title:       constants_steampipe.CmdExit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		constants_steampipe.CmdQuit: {
			title:       constants_steampipe.CmdQuit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		constants_steampipe.CmdTableList: {
			title:       constants_steampipe.CmdTableList,
			handler:     listTables,
			validator:   atMostNArgs(1),
			description: "List or describe tables",
		},
		constants_steampipe.CmdSeparator: {
			title:       constants_steampipe.CmdSeparator,
			handler:     setViperConfigFromArg(constants.ArgSeparator),
			validator:   exactlyNArgs(1),
			description: "Set csv output separator",
		},
		constants_steampipe.CmdHeaders: {
			title:       "headers",
			handler:     setHeader,
			validator:   booleanValidator(constants_steampipe.CmdHeaders, validatorFromArgsOf(constants_steampipe.CmdHeaders)),
			description: "Enable or disable column headers",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on headers in output"},
				{value: constants.ArgOff, description: "Turn off headers in output"},
			},
			completer: completerFromArgsOf(constants_steampipe.CmdHeaders),
		},
		constants_steampipe.CmdMulti: {
			title:       "multi-line",
			handler:     setMultiLine,
			validator:   booleanValidator(constants_steampipe.CmdMulti, validatorFromArgsOf(constants_steampipe.CmdMulti)),
			description: "Enable or disable multiline mode",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on multiline mode"},
				{value: constants.ArgOff, description: "Turn off multiline mode"},
			},
			completer: completerFromArgsOf(constants_steampipe.CmdMulti),
		},
		constants_steampipe.CmdTiming: {
			title:       "timing",
			handler:     setTiming,
			validator:   booleanValidator(constants_steampipe.CmdTiming, validatorFromArgsOf(constants_steampipe.CmdTiming)),
			description: "Enable or disable query execution timing",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Display time elapsed after every query"},
				{value: constants.ArgOff, description: "Turn off query timer"},
			},
			completer: completerFromArgsOf(constants_steampipe.CmdTiming),
		},
		constants_steampipe.CmdOutput: {
			title:       constants_steampipe.CmdOutput,
			handler:     setViperConfigFromArg(constants.ArgOutput),
			validator:   composeValidator(exactlyNArgs(1), validatorFromArgsOf(constants_steampipe.CmdOutput)),
			description: "Set output format: csv, json, table or line",
			args: []metaQueryArg{
				{value: constants_steampipe.OutputFormatJSON, description: "Set output to JSON"},
				{value: constants_steampipe.OutputFormatCSV, description: "Set output to CSV"},
				{value: constants_steampipe.OutputFormatTable, description: "Set output to Table"},
				{value: constants_steampipe.OutputFormatLine, description: "Set output to Line"},
			},
			completer: completerFromArgsOf(constants_steampipe.CmdOutput),
		},
		constants_steampipe.CmdCache: {
			title:       constants_steampipe.CmdCache,
			handler:     cacheControl,
			validator:   validatorFromArgsOf(constants_steampipe.CmdCache),
			description: "Enable, disable or clear the query cache",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on caching"},
				{value: constants.ArgOff, description: "Turn off caching"},
				{value: constants.ArgClear, description: "Clear the cache"},
			},
			completer: completerFromArgsOf(constants_steampipe.CmdCache),
		},
		constants_steampipe.CmdCacheTtl: {
			title:       constants_steampipe.CmdCacheTtl,
			handler:     cacheTTL,
			validator:   atMostNArgs(1),
			description: "Set the cache ttl (time-to-live)",
		},
		constants_steampipe.CmdInspect: {
			title:   constants_steampipe.CmdInspect,
			handler: inspect,
			// .inspect only supports a single arg, however the arg validation code cannot understand escaped arguments
			// e.g. it will treat csv."my table" as 2 args
			// the logic to handle this escaping is lower down so we just validate to ensure at least one argument has been provided
			validator:   atLeastNArgs(0),
			description: "View connections, tables & column information",
			completer:   inspectCompleter,
		},
		constants_steampipe.CmdConnections: {
			title:       constants_steampipe.CmdConnections,
			handler:     listConnections,
			validator:   noArgs,
			description: "List active connections",
		},
		constants_steampipe.CmdClear: {
			title:       constants_steampipe.CmdClear,
			handler:     clearScreen,
			validator:   noArgs,
			description: "Clear the console",
		},
		constants_steampipe.CmdSearchPath: {
			title:       constants_steampipe.CmdSearchPath,
			handler:     setOrGetSearchPath,
			validator:   atMostNArgs(1),
			description: "Display the current search path, or set the search-path by passing in a comma-separated list",
		},
		constants_steampipe.CmdSearchPathPrefix: {
			title:       constants_steampipe.CmdSearchPathPrefix,
			handler:     setSearchPathPrefix,
			validator:   exactlyNArgs(1),
			description: "Set a prefix to the current search-path",
		},
		constants_steampipe.CmdAutoComplete: {
			title:       "auto-complete",
			handler:     setAutoComplete,
			validator:   booleanValidator(constants_steampipe.CmdAutoComplete, validatorFromArgsOf(constants_steampipe.CmdAutoComplete)),
			description: "Enable or disable auto-completion",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on auto-completion"},
				{value: constants.ArgOff, description: "Turn off auto-completion"},
			},
			completer: completerFromArgsOf(constants_steampipe.CmdAutoComplete),
		},
	}
}
