mod "bad_mod_with_sp_version_require_not_met" {
  title       = "Bad Mod 2"
  description = "This mod is used to test that the steampipe commands always respect the requirements mentioned in mod.sp require section"

  require {
    steampipe = "10.99.99"
  }
}
