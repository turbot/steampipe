mod "dep1" {
  requires {
    mod "github.com/kaidaguerre/steampipe-mod-m1" {
      version = "v1.*"
    }
    mod "github.com/kaidaguerre/steampipe-mod-m2" {
      version = "v3.0"
    }
    mod "github.com/turbot/steampipe-mod-aws-compliance" {
      version = "0"
    }
  }
}
