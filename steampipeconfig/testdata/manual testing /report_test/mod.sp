
variable mandatory_tags{
  default = "FOO"
}

mod "foo"{
  title = "FOO"
  description = "THIS IS M1"
  require {
    steampipe = "0.13.1"
    plugin "aws" {
      version = "0.54.0"
    }
    mod "github.com/pskrbasu/steampipe-mod-m1" {
      version = "latest"
      args = {
        mandatory_tags = var.mandatory_tags
      }
    }
  }
}