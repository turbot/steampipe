package constants

const DefaultSPCContent = `
# options "connection" {
#     cache           = "on"   # on, off
#     cache_ttl       = 300    # int = time in seconds
# }
# options "database" {
#     port   = 9193
#     listen = "local"
# }
# options "terminal" {
#     header      = "off"      # on, off
#     multi       = "off"      # on, off
#     output      = "table"    # json, csv, table, line
#     separator   = ","        # any single char
#     timing      = "off"      # on, off
# }
# options "general" {
#     update_check  = "on"     # on, off
# }

`
