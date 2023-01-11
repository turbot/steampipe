variable "sample_var_1"{
	type = string
	default = "steampipe_var"
}


query "sample_query_1"{
	title ="Sample query 1"
	description = "query 1 - 3 params all with defaults"
	sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, $2::text, $3::text) as reason"
	param "p1"{
			description = "p1"
			default = var.sample_var_1
	}
	param "p2"{
			description = "p2"
			default = "because_def "
	}
	param "p3"{
			description = "p3"
			default = "string"
	}
}

control "sample_control_1" {
  title = "Sample control 1"
  description = "Sample control to test introspection functionality"
  query = query.sample_query_1
  severity = "high"
}

benchmark "sample_benchmark_1" {
	title = "Sample benchmark 1"
	description = "Sample benchmark to test introspection functionality"
	children = [
		control.sample_control_1
	]
}

