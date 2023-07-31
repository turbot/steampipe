workspace "default" {
  cloud_host = "latestpipe.turbot.io/"
  cloud_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_default"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
  search_path =  ""
  search_path_prefix = "abc"
  watch = false
  introspection = "info"
  query_timeout = 180
  max_parallel = 1
  install_dir = "sp_install_dir_default"
  theme = "plain"
  progress = true
  input = false
  options "query" {
    autocomplete = false
    header = false
    multi = true
    output = "json"
    separator = "|"
    timing = true
  }
  options "check" {
    header = false
    output = "json"
    separator = "|"
    timing = true
  }
  options "dashboard" {
    browser = true
  }
}

workspace "sample" {
  cloud_host = "testpipe.turbot.io/"
  cloud_token = "spt_012faketoken34567890_012faketoken3456789099998"
  mod_location = "sp_install_dir_sample"
  snapshot_location = "snaps2"
  workspace_database = "fk43e6"
  search_path =  "abc,def"
  search_path_prefix = "abc"
  watch = true
  introspection = "control"
  query_timeout = 200
  max_parallel = 2
  install_dir = "sp_install_dir_sample"
  theme = "dark"
  progress = false
  input = true
  options "query" {
    autocomplete = true
    header = true
    multi = false
    output = "csv"
    separator = ","
    timing = false
  }
  options "check" {
    header = true
    output = "csv"
    separator = ","
    timing = false
  }
  options "dashboard" {
    browser = false
  }
}