package constants

const DefaultSPCContent = `
# options "connection" {
#   cache     = true # true, false
#   cache_ttl = 300  # expiration (TTL) in seconds
# }

# options "database" {
#   port        = 9193    # any valid, open port number
#   listen      = "local" # local, network
#   search_path =  ""     # comma-separated string; an exact search_path
# }

# options "terminal" {
#   multi               = false   # true, false
#   output              = "table" # json, csv, table, line
#   header              = true    # true, false
#   separator           = ","     # any single char
#   timing      	    = false   # true, false
#   search_path         =  ""     # comma-separated string; an exact search_path
#   search_path_prefix  =  ""     # comma-separated string; a search_path_prefix to prepend to the search_path
# }

# options "general" {
#   update_check = true # true, false
# }
`
