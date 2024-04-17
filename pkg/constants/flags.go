package constants

import (
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/pipe-fittings/constants"
)

type QueryOutputMode enumflag.Flag

const (
	QueryOutputModeCsv QueryOutputMode = iota
	QueryOutputModeJson
	QueryOutputModeLine
	QueryOutputModeSnapshot
	QueryOutputModeSnapshotShort
	QueryOutputModeTable
)

var QueryOutputModeIds = map[QueryOutputMode][]string{
	QueryOutputModeCsv:           {constants.OutputFormatCSV},
	QueryOutputModeJson:          {constants.OutputFormatJSON},
	QueryOutputModeLine:          {constants.OutputFormatLine},
	QueryOutputModeSnapshot:      {constants.OutputFormatSnapshot},
	QueryOutputModeSnapshotShort: {constants.OutputFormatSnapshotShort},
	QueryOutputModeTable:         {constants.OutputFormatTable},
}

type QueryTimingMode enumflag.Flag

const (
	QueryTimingModeOff QueryTimingMode = iota
	QueryTimingModeOn
	QueryTimingModeVerbose
)

var QueryTimingModeIds = map[QueryTimingMode][]string{
	QueryTimingModeOff:     {constants.ArgOff},
	QueryTimingModeOn:      {constants.ArgOn},
	QueryTimingModeVerbose: {constants.ArgVerbose},
}

var QueryTimingValueLookup = map[string]struct{}{
	constants.ArgOff:     {},
	constants.ArgOn:      {},
	constants.ArgVerbose: {},
}

type CheckTimingMode enumflag.Flag

const (
	CheckTimingModeOff CheckTimingMode = iota
	CheckTimingModeOn
)

var CheckTimingModeIds = map[CheckTimingMode][]string{
	CheckTimingModeOff: {constants.ArgOff},
	CheckTimingModeOn:  {constants.ArgOn},
}

var CheckTimingValueLookup = map[string]struct{}{
	constants.ArgOff: {},
	constants.ArgOn:  {},
}

type CheckOutputMode enumflag.Flag

const (
	CheckOutputModeText  CheckOutputMode = iota
	CheckOutputModeBrief CheckOutputMode = iota
	CheckOutputModeCsv
	CheckOutputModeHTML
	CheckOutputModeJSON
	CheckOutputModeMd
	CheckOutputModeSnapshot
	CheckOutputModeSnapshotShort
	CheckOutputModeNone
)

var CheckOutputModeIds = map[CheckOutputMode][]string{
	CheckOutputModeText:          {constants.OutputFormatText},
	CheckOutputModeBrief:         {constants.OutputFormatBrief},
	CheckOutputModeCsv:           {constants.OutputFormatCSV},
	CheckOutputModeHTML:          {constants.OutputFormatHTML},
	CheckOutputModeJSON:          {constants.OutputFormatJSON},
	CheckOutputModeMd:            {constants.OutputFormatMD},
	CheckOutputModeSnapshot:      {constants.OutputFormatSnapshot},
	CheckOutputModeSnapshotShort: {constants.OutputFormatSnapshotShort},
	CheckOutputModeNone:          {constants.OutputFormatNone},
}

func FlagValues[T comparable](mappings map[T][]string) []string {
	var res = make([]string, 0, len(mappings))
	for _, v := range mappings {
		res = append(res, v[0])
	}
	return res

}
