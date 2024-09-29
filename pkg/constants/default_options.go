package constants

// DefaultConnectionConfigContent is the content of the sample connection config file(default.spc.sample),
// that is created if it does not exist
const DefaultConnectionConfigContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/cli-args
#

# options "database" {
#   port               = 9193                  # any valid, open port number
#   listen             = "local"               # local (alias for localhost), network (alias for *), or a comma separated list of hosts and/or IP addresses , or any valid combination of hosts and/or IP addresses
#   search_path        = "aws,aws2,gcp,gcp2"   # comma-separated string; an exact search_path
#   search_path_prefix = "aws"                 # comma-separated string; a search_path prefix
#   start_timeout      = 30                    # maximum time (in seconds) to wait for the database to start up
#   cache              = true                  # true, false
#   cache_max_ttl      = 900                   # max expiration (TTL) in seconds
#   cache_max_size_mb  = 1024                  # max total size of cache across all plugins
# }

# options "general" {
#   update_check = true    		# true, false
#   telemetry    = "info"  		# info, none
#   log_level    = "info"  		# trace, debug, info, warn, error
#   memory_max_mb    = "1024"	# the maximum memory to allow the CLI process in MB 
# }

# options "plugin" {
#   memory_max_mb    = "1024"	# the default maximum memory to allow a plugin process - used if there is not max memory specified in the 'plugin' block' for that plugin
#   start_timeout    = 30       # maximum time (in seconds) to wait for a plugin to start up
# }
`
