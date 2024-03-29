
workspace "default" {
  introspection = "info"
  pipes_host = "latestpipe.turbot.io/"
  pipes_token = "spt_012faketoken34567890_012faketoken3456789099999"
  install_dir = "sp_install_dir_default"
  mod_location = "sp_install_dir_default"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
}

workspace "sample" {
  introspection = "control"
  pipes_host = "testpipe.turbot.io"
  pipes_token = "spt_012faketoken34567890_012faketoken3456789099999"
  install_dir = "sp_install_dir_sample"
  mod_location = "sp_install_dir_sample"
  snapshot_location = "snap"
  workspace_database = "fk43e8"
}