mod "local" {
  title = "dep_empty"
  require {
    mod "github.com/kaidaguerre/steampipe-mod-m2" {
      version = "latest"
    }
    mod "github.com/turbot/steampipe-mod-aws-compliance" {
      version = "latest"
    }
  }
}
