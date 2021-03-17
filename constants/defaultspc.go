package constants

const DefaultSPCContent = `
# options "connection" {
#     cache           = "true"     # true, false
#     cache_ttl       = 300        # int = time in seconds
# }
# options "database" {
#     port   = 9193
#     listen = "local"
# }
# options "console" {
#     header      = "off"      # on, off
#     multi       = "off"     # on, off
#     output      = "table"   # json, csv, table, line
#     separator   = ","       # any single char
#     timing      = "off"     # on, off
# }
# options "general" {
#     log_level  = "warn"     # trace, debug, info, warn, error
# }

`
