
options "connection" {
   cache     = true # true, false
   cache_ttl = 300  # expiration (TTL) in seconds
 }

options "database" {
  port   = 9193    # any valid, open port number
  listen = "local" # local, network
}

options "terminal" {
  multi     = false   # true, false
  output    = "table" # json, csv, table, line
  header    = true    # true, false
  separator = ","     # any single char
  timing    = false   # true, false
  search_path    = "aws,gcp"
}

options "general" {
   update_check = true # true, false
}
