mod "local" {
  title = "dependent_mod"
  require {
    mod "github.com/pskrbasu/steampipe-mod-top-level" {
      version = "3.0.0"
    }
  }
}
