mod "mod_with_both_version_and_minversion_in_plugin_block" {
  title = "mod_with_both_version_and_minversion_in_plugin_block"
  require {
    plugin "chaos" {
      version = "0.1.0"
      min_version = "0.1.0"
    }
  }
}