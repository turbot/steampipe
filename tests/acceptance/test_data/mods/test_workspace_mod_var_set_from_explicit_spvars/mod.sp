mod "test_vars_workspace_mod" {
  title = "test_vars_workspace_mod"
}

query "version" {
  sql = "select $1::text as reason, $1::text as resource, 'ok' as status"
  param "p1"{
    description = "p1"
    default = var.version
	}
}

variable "version"{
	type = string
	default = "v2.0.0"
}