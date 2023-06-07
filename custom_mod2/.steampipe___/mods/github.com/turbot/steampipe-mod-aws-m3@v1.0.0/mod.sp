
variable "v1" {
  type        = string
  default = "m3-default"
}


mod "m3" {
  # hub metadata
  title         = "Mod 3"
}

dashboard "d1"{
    chart "c1"{
        query = "select 1"
    }
}