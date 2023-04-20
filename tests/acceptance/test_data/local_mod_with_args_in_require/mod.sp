mod "local_mod_with_args_in_require" {
  require {
    mod "github.com/pskrbasu/steampipe-mod-dependency-vars-1" {
      version = "*"
      args = {
        version: var.top
      }
    }
  }
}

variable "top" {
  default = "v3.0.0"
}
