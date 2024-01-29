mod "test_vars_dependency_mod" {
  title = "test_vars_dependency_mod"
  require {
    mod "github.com/pskrbasu/steampipe-mod-dependency-vars-1" {
      version = "*"
    }
  }
}
