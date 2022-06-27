# invalid for workspace
options "database" {
  port   = 9193    # any valid, open port number
  listen = "local" # local, network
  search_path    = "aws,gcp,foo"
}