options "database" {
  port                = 9193
  listen              = "local"
  search_path         = "abc,def"
  search_path_prefix  = "def"
  start_timeout       = 30
  cache               = true
  cache_max_ttl       = 900
  cache_max_size_mb   = 1024
}

options "dashboard" {
  port   = 9193
  listen = "local"
}

options "general" {
  update_check = true
  telemetry    = "info"
  max_parallel = 1
}