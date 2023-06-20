mod "local" {
  title = "custom_mod"

  require {
    mod "github.com/kaidaguerre/steampipe-mod-m1" {
      version = "*"
      args = {
          v1 = "top level arg"
      }
    }
    mod "github.com/kaidaguerre/steampipe-mod-m2" {
      version = "*"
    }
    mod "github.com/kaidaguerre/steampipe-mod-m3" {
      version = "*"
    }
  }
}