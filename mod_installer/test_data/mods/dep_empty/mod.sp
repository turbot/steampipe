mod "local" {
  title = "dep_empty"
  requires {
    mod "github.com/turbot/steampipe-mod-aws-compliance" {
      version = "0"
    }
  }
}
