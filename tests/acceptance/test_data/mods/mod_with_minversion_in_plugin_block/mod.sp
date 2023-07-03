mod "mod_with_minversion_in_plugin_block" {
  title = "mod_with_minversion_in_plugin_block"
  require {
    plugin "chaos" {
      min_version = "0.1.0"
    }
  }
}