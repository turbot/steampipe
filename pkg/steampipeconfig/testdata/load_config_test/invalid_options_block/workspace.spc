# invalid for workspace
options "database" {
  port        = 9193    # any valid, open port number
  listen      = "local" # local (alias for localhost), network (alias for *), or a comma separated list of hosts and/or IP addresses
  search_path = "aws,gcp,foo"
}
