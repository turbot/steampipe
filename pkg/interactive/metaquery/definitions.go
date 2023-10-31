package metaquery

import (
	"github.com/turbot/pipe-fittings/constants"
	localconstants "github.com/turbot/steampipe/pkg/constants"
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
		localconstants.CmdHelp: {
			title:       localconstants.CmdHelp,
			handler:     doHelp,
			validator:   noArgs,
			description: "Show steampipe help",
		},
		localconstants.CmdExit: {
			title:       localconstants.CmdExit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		localconstants.CmdQuit: {
			title:       localconstants.CmdQuit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		localconstants.CmdTableList: {
			title:       localconstants.CmdTableList,
			handler:     listTables,
			validator:   atMostNArgs(1),
			description: "List or describe tables",
		},
		localconstants.CmdSeparator: {
			title:       localconstants.CmdSeparator,
			handler:     setViperConfigFromArg(constants.ArgSeparator),
			validator:   exactlyNArgs(1),
			description: "Set csv output separator",
		},
		localconstants.CmdHeaders: {
			title:       "headers",
			handler:     setHeader,
			validator:   booleanValidator(localconstants.CmdHeaders, validatorFromArgsOf(localconstants.CmdHeaders)),
			description: "Enable or disable column headers",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on headers in output"},
				{value: constants.ArgOff, description: "Turn off headers in output"},
			},
			completer: completerFromArgsOf(localconstants.CmdHeaders),
		},
		localconstants.CmdMulti: {
			title:       "multi-line",
			handler:     setMultiLine,
			validator:   booleanValidator(localconstants.CmdMulti, validatorFromArgsOf(localconstants.CmdMulti)),
			description: "Enable or disable multiline mode",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on multiline mode"},
				{value: constants.ArgOff, description: "Turn off multiline mode"},
			},
			completer: completerFromArgsOf(localconstants.CmdMulti),
		},
		localconstants.CmdTiming: {
			title:       "timing",
			handler:     setTiming,
			validator:   booleanValidator(localconstants.CmdTiming, validatorFromArgsOf(localconstants.CmdTiming)),
			description: "Enable or disable query execution timing",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Display time elapsed after every query"},
				{value: constants.ArgOff, description: "Turn off query timer"},
			},
			completer: completerFromArgsOf(localconstants.CmdTiming),
		},
		localconstants.CmdOutput: {
			title:       localconstants.CmdOutput,
			handler:     setViperConfigFromArg(constants.ArgOutput),
			validator:   composeValidator(exactlyNArgs(1), validatorFromArgsOf(localconstants.CmdOutput)),
			description: "Set output format: csv, json, table or line",
			args: []metaQueryArg{
				{value: constants.OutputFormatJSON, description: "Set output to JSON"},
				{value: constants.OutputFormatCSV, description: "Set output to CSV"},
				{value: constants.OutputFormatTable, description: "Set output to Table"},
				{value: constants.OutputFormatLine, description: "Set output to Line"},
			},
			completer: completerFromArgsOf(localconstants.CmdOutput),
		},
		localconstants.CmdCache: {
			title:       localconstants.CmdCache,
			handler:     cacheControl,
			validator:   validatorFromArgsOf(localconstants.CmdCache),
			description: "Enable, disable or clear the query cache",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on caching"},
				{value: constants.ArgOff, description: "Turn off caching"},
				{value: constants.ArgClear, description: "Clear the cache"},
			},
			completer: completerFromArgsOf(localconstants.CmdCache),
		},
		localconstants.CmdCacheTtl: {
			title:       localconstants.CmdCacheTtl,
			handler:     cacheTTL,
			validator:   atMostNArgs(1),
			description: "Set the cache ttl (time-to-live)",
		},
		localconstants.CmdInspect: {
			title:   localconstants.CmdInspect,
			handler: inspect,
			// .inspect only supports a single arg, however the arg validation code cannot understand escaped arguments
			// e.g. it will treat csv."my table" as 2 args
			// the logic to handle this escaping is lower down so we just validate to ensure at least one argument has been provided
			validator:   atLeastNArgs(0),
			description: "View connections, tables & column information",
			completer:   inspectCompleter,
		},
		localconstants.CmdConnections: {
			title:       localconstants.CmdConnections,
			handler:     listConnections,
			validator:   noArgs,
			description: "List active connections",
		},
		localconstants.CmdClear: {
			title:       localconstants.CmdClear,
			handler:     clearScreen,
			validator:   noArgs,
			description: "Clear the console",
		},
		localconstants.CmdSearchPath: {
			title:       localconstants.CmdSearchPath,
			handler:     setOrGetSearchPath,
			validator:   atMostNArgs(1),
			description: "Display the current search path, or set the search-path by passing in a comma-separated list",
		},
		localconstants.CmdSearchPathPrefix: {
			title:       localconstants.CmdSearchPathPrefix,
			handler:     setSearchPathPrefix,
			validator:   exactlyNArgs(1),
			description: "Set a prefix to the current search-path",
		},
		localconstants.CmdAutoComplete: {
			title:       "auto-complete",
			handler:     setAutoComplete,
			validator:   booleanValidator(localconstants.CmdAutoComplete, validatorFromArgsOf(localconstants.CmdAutoComplete)),
			description: "Enable or disable auto-completion",
			args: []metaQueryArg{
				{value: constants.ArgOn, description: "Turn on auto-completion"},
				{value: constants.ArgOff, description: "Turn off auto-completion"},
			},
			completer: completerFromArgsOf(localconstants.CmdAutoComplete),
		},
	}
}
