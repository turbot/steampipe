package constants

// DefaultConnectionConfigContent is the content of the default connection config file, default.spc,
// that is created if it does not exist
const DefaultConnectionConfigContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/cli-args
#

# options "connection" {
#   cache     = true # true, false
#   cache_ttl = 300  # expiration (TTL) in seconds
# }

# options "database" {
#   port          = 9193    # any valid, open port number
#   listen        = "local" # local, network
#   search_path   =  ""     # comma-separated string
#   start_timeout = 30      # maximum time it should take for the database service to start accepting queries (in seconds)
# }

# options "terminal" {
#   multi               = false   # true, false
#   output              = "table" # json, csv, table, line
#   header              = true    # true, false
#   separator           = ","     # any single char
#   timing              = false   # true, false
#   search_path         =  ""     # comma-separated string
#   search_path_prefix  =  ""     # comma-separated string
#   watch  			        =  true   # true, false
#   autocomplete        =  true   # true, false
# }

# options "general" {
#   update_check = true # true, false
# }
`
