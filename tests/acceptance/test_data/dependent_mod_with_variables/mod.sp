mod "local" {
  title = "dependent_mod"
  require {
    mod "github.com/pskrbasu/steampipe-mod-m1" {
      version = "4.0"
      args = {
        dep_mod_var2: "select 'dep_mod_var2_set_in_mod_require' as a"
      }
    }
  }
}
