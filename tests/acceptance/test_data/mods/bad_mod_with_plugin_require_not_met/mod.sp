mod "bad_mod_with_require_not_met" {
  title       = "Bad Mod"
  description = "This mod is used to test that the steampipe commands always respect the requirements mentioned in mod.sp require section"

  require {
    plugin "gcp" {
      version = "99.21.0"
    }
  }
}
