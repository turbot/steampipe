mod "local_mod_with_args_in_require" {
  require {
    mod "github.com/pskrbasu/steampipe-mod-dependency-vars-1" {
      version = "*"
    }
  }
}