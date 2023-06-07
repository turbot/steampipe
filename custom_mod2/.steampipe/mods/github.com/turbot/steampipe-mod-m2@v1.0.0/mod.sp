
variable "v1" {
  type        = string
  default = "m2-default"
}


mod "m2" {
  # hub metadata
  title         = "Mod 2"

  require {
    mod "github.com/turbot/steampipe-mod-m3" {
      version = "*"
    }
  }
}

dashboard "d1"{
    chart "c1"{
        query = "select 1"
    }
}