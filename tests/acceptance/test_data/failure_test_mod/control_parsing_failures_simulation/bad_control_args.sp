benchmark "control_parsing_failures_simulation" {
  title         = "Benchmark to simulate parsing failures for controls in steampipe(WILL FAIL)"
  children = [
    control.control_fail_with_no_query_no_sql,
    control.control_fail_with_both_query_and_sql,
    control.control_fail_with_params_and_query,
    control.control_fail_with_query_with_no_def_and_named_args_passed,
    control.control_fail_with_insufficient_positional_args_passed,
    control.control_fail_with_insufficient_named_args_passed
  ]
}

control "control_fail_with_no_query_no_sql" {
  title = "Control to simulate parsing failure for control(no query, no sql)"
  description = "A control must define either a 'sql' property or a 'query' property"
}

control "control_fail_with_both_query_and_sql" {
  title = "Control to simulate parsing failure for control(both query and sql)"
  description = "A control must define either a 'sql' property or a 'query' property, not both"
  query = query.query_params_with_all_defaults
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
}

control "control_fail_with_params_and_query" {
  title = "Control to simulate parsing failure for control(control contains params)"
  description = "Control has query property set so cannot define param blocks"
  query = query.query_params_with_all_defaults
  param "p1"{
    description = "First parameter"
    default = "default_parameter_1"
  }
  param "p2"{
    description = "Second parameter"
    default = "default_parameter_2"
  }
  param "p3"{
    description = "Third parameter"
    default = "default_parameter_3"
  }
}

control "control_fail_with_query_with_no_def_and_named_args_passed" {
  title = "Control to simulate parsing failure for control(control refers to a query with no param definitions and some named arguments passed)"
  description = "Control referring to a query with no param definitions"
  query = query.query_with_no_param_defs
  args = {
    "p1" = "command_parameter_1"
    "p2" = "command_parameter_2"
    "p3" = "command_parameter_3"
  }
}

control "control_fail_with_insufficient_positional_args_passed" {
  title = "Control fail with insufficient positional args passed"
  description = "Control to simulate parsing failure for control(control refers to a query with no param defaults and partial positional arguments passed)"
  query = query.query_with_param_defs_no_defaults
  args = [ "command_argument_1", "command_argument_2" ]
}

control "control_fail_with_insufficient_named_args_passed" {
  title = "Control fail with insufficient positional args passed"
  description = "Control to simulate parsing failure for control(control refers to a query with no param defaults and partial positional arguments passed)"
  query = query.query_with_param_defs_no_defaults
  args = {
    "p1" = "command_parameter_1"
    "p2" = "command_parameter_2"
  }
}