query "version" {
  sql = "select $1::text as reason, $1::text as resource, 'ok' as status"
  param "p1"{
    description = "p1"
    default = var.version
	}
}

variable "version"{
	type = string
}