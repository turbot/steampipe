mod "mod_with_old_steampipe_and_new_steampipe_block_in_require" {
  title = "mod_with_old_steampipe_and_new_steampipe_block_in_require"
  require {
    steampipe = "0.18.0"
    steampipe {
      min_version = "0.18.0"
    }
  }
}
