mod "bad_mod_with_dep_mod_version_require_not_met" {
  title       = "Bad Mod 3"
  description = "This mod is used to test that the steampipe commands always respect the requirements mentioned in mod.sp require section"

  require {
    mod "github.com/turbot/steampipe-mod-aws-compliance" {
      version = "99.21.0"
    }
  }
}
