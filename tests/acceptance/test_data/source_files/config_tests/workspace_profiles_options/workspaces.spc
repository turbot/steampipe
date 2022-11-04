workspace "default" {
  cloud_host = "latestpipe.turbot.io/"
  cloud_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_default"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
  options "terminal" {
    multi               = true
    output              = "json"
    header              = false
    separator           = "|"
    timing              = true
    search_path         =  ""
    search_path_prefix  =  "abc"
    watch  			    =  false
    autocomplete       =  false
  }
  options "general" {
    update_check = false
  }
}

workspace "sample" {
  cloud_host = "latestpipe.turbot.io/"
  cloud_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_sample"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
  options "terminal" {
    multi               = true
    output              = "csv"
    header              = false
    separator           = ";"
    timing              = true
    search_path         =  "abc"
    search_path_prefix  =  "abc, def"
    watch  			    =  false
    autocomplete       =  true
  }
  options "general" {
    update_check = true
  }
}