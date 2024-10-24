package constants

// Metaquery commands

const (
	CmdTableList        = ".tables"             // List all tables
	CmdOutput           = ".output"             // Set output mode
	CmdTiming           = ".timing"             // Toggle query timer
	CmdHeaders          = ".header"             // Toggle headers output
	CmdSeparator        = ".separator"          // Set the column separator
	CmdExit             = ".exit"               // Exit the interactive prompt
	CmdQuit             = ".quit"               // Alias for .exit
	CmdInspect          = ".inspect"            // inspect
	CmdConnections      = ".connections"        // list all connections
	CmdMulti            = ".multi"              // toggle multi line query
	CmdClear            = ".clear"              // clear the console
	CmdHelp             = ".help"               // list all meta commands
	CmdSearchPath       = ".search_path"        // Set or show search-path
	CmdSearchPathPrefix = ".search_path_prefix" // set search path prefix
	CmdCache            = ".cache"              // cache control
	CmdCacheTtl         = ".cache_ttl"          // set cache ttl
	CmdAutoComplete     = ".autocomplete"       // enable or disable auto complete
)
