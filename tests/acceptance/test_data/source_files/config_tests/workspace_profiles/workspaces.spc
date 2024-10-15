
workspace "default" {
  pipes_host = "latestpipe.turbot.io/"
  pipes_token = "spt_012faketoken34567890_012faketoken3456789099999"
  install_dir = "sp_install_dir_default"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
}

workspace "sample" {
  pipes_host = "testpipe.turbot.io"
  pipes_token = "spt_012faketoken34567890_012faketoken3456789099999"
  install_dir = "sp_install_dir_sample"
  snapshot_location = "snap"
  workspace_database = "fk43e8"
}