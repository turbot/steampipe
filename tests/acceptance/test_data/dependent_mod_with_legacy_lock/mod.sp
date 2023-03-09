mod "local" {
  title = "dependent_mod"
  require {
    mod "github.com/pskrbasu/steampipe-mod-m1" {
      version = "4.0"
    }
  }
}
