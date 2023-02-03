options "connection" {
  cache     = true # true, false
  cache_ttl = 300  # expiration (TTL) in seconds
}

options "terminal" {
  multi               = false   # true, false
  output              = "table" # json, csv, table, line
  header              = true    # true, false
  separator           = ","     # any single char
  timing              = false   # true, false
  search_path         =  ""     # comma-separated string
  search_path_prefix  =  ""     # comma-separated string
  watch  			        =  true   # true, false
  autocomplete        =  true   # true, false
}

options "general" {
  update_check = false # true, false
}
