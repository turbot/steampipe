mod "foo"{
  title = "FOO"
  description = "THIS IS M1"
  require {
    mod "github.com/turbot/steampipe-mod-aws-compliance" {
      version = "latest"
    }
  }
}