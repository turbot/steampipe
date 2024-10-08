workspace "default" {
  pipes_host = "latestpipe.turbot.io/"
  pipes_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_default"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
  search_path =  ""
  search_path_prefix = "abc"
  watch = false
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
}

workspace "sample" {
  pipes_host = "latestpipe.turbot.io/"
  pipes_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_sample"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
  search_path = "abc"
  search_path_prefix = "abc, def"
  watch = false
  options "query" {
    autocomplete =  true
    header = false
    multi = true
    output = "csv"
    separator = ";"
    timing = true
  }
  options "check" {
    header = false
    output = "csv"
    separator = ";"
    timing = true
  }
}