package constants

const DefaultSPCContent = `
# options "connection" {
#     cache           = true   # true, false
#     cache_ttl       = 300    # int = time in seconds
# }

# options "database" {
#     port   = 9193
#     listen = "local"
# }

# options "terminal" {
#     multi       = false      # true, false
#     output      = "table"    # json, csv, table, line
#     header      = false      # true, false
#     separator   = ","        # any single char
#     timing      = false      # true, false
# }

# options "general" {
#     update_check  = true     # true, false
# }

`
