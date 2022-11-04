
workspace "default" {
  cloud_host = "latestpipe.turbot.io/"
  cloud_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_default"
  snapshot_location = "snaps"
  workspace_database = "fk43e7"
}

workspace "sample" {
  cloud_host = "testpipe.turbot.io"
  cloud_token = "spt_012faketoken34567890_012faketoken3456789099999"
  mod_location = "sp_install_dir_sample"
  snapshot_location = "snap"
  workspace_database = "fk43e8"
}