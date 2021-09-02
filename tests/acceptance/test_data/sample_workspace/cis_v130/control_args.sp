benchmark "query_and_control_parameters_benchmark" {
  title         = "Benchmark to test the query and control parameter functionalities in steampipe"
  children = [
    control.query_params_with_defaults_and_no_args,
    control.query_params_with_defaults_and_some_args,
    control.query_params_with_no_defaults_and_no_args,
    control.query_params_with_no_defaults_with_args,
    control.query_params_array_with_default,
    control.query_params_map_with_default,
    control.query_params_invalid_arg_syntax
  ]
}

control "query_params_with_defaults_and_no_args" {
  title = "Control to test query param functionality with defaults(and no args passed)"
  query = query.query_params_with_all_defaults
}

control "query_params_with_defaults_and_some_args" {
  title = "Control to test query param functionality with defaults(and some args passed in query)"
  query = query.query_params_with_all_defaults
  args = {
    "p2" = "command_parameter_2 "
  }
}

control "query_params_with_no_defaults_and_no_args" {
  title = "Control to test query param functionality with no defaults(and no args passed)"
  query = query.query_params_with_no_defaults
}

control "query_params_with_no_defaults_with_args" {
  title = "Control to test query param functionality with no defaults(and args passed in query)"
  query = query.query_params_with_no_defaults
  args = {
    "p1" = "command_parameter_1 "
    "p2" = "command_parameter_2 "
    "p3" = "command_parameter_3"
  }
}

control "query_params_array_with_default" {
  title = "Control to test query param functionality with an array param with default(and no args passed)"
  query = query.query_array_params_with_default
}

control "query_params_map_with_default" {
  title = "Control to test query param functionality with a map param with default(and no args passed)"
  query = query.query_map_params_with_default
}

control "query_params_invalid_arg_syntax" {
  title = "Control to test query param functionality with a map param with no default(and invalid args passed in query)"
  query = query.query_map_params_with_no_default
  args = {
    "p1" = "command_parameter_1"
  }
}

control "control_args_array_with_no_defaults" {
  title = "Control to test the control args functionality with no defaults(arguments passed)"
  query = query.query_params_with_all_defaults
}

