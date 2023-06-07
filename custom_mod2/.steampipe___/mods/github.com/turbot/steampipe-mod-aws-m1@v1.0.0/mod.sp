
variable "v1" {
  type        = string
  default = "m1-default"
}


mod "m1" {
  # hub metadata
  title         = "Mod 1"

  require {
    mod "github.com/turbot/steampipe-mod-m2" {
      version = "*"
    }
  }
}

dashboard "d1"{
    chart "c1"{
        query = "select 1"
    }
}