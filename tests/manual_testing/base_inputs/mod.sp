mod "reports_poc" {
  title = "Reports POC"
  require {
    mod "github.com/turbot/steampipe-mod-aws-compliance" {
      version = "latest"
    }
  }
}
