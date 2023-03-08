
options "connection" {
  cache     = true # true, false
  cache_ttl = 300  # expiration (TTL) in seconds
}

options "database" {
  port        = 9193    # any valid, open port number
  listen      = "local" # local (alias for localhost), network (alias for *), or a comma separated list of hosts and/or IP addresses
  search_path = "aws,gcp,foo"
}

options "terminal" {
  multi        = false   # true, false
  output       = "table" # json, csv, table, line
  header       = true    # true, false
  separator    = ","     # any single char
  timing       = false   # true, false
  search_path  = "aws,gcp"
  autocomplete = "true"
}

options "general" {
  update_check = true # true, false
}
