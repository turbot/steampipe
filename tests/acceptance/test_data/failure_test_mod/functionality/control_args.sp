benchmark "query_and_control_parameters_benchmark" {
  title         = "Benchmark to test the query and control parameter functionalities in steampipe"
  children = [
    control.query_params_with_defaults_and_no_args,
    control.query_params_with_defaults_and_partial_named_args,
    control.query_params_with_defaults_and_partial_positional_args,
    control.query_params_with_defaults_and_all_named_args,
    control.query_params_with_defaults_and_all_positional_args,
    control.query_params_with_no_defaults_and_no_args,
    control.query_params_with_no_defaults_with_named_args,
    control.query_params_with_no_defaults_with_positional_args,
    control.query_params_array_with_default,
    control.query_params_map_with_default,
    control.query_params_invalid_arg_syntax,
    control.query_inline_sql_from_control_with_partial_named_args,
    control.query_inline_sql_from_control_with_partial_positional_args,
    control.query_inline_sql_from_control_with_no_args,
    control.query_inline_sql_from_control_with_all_positional_args,
    control.query_inline_sql_from_control_with_all_named_args
  ]
}

control "query_params_with_defaults_and_no_args" {
  title = "Control to test query param functionality with defaults(and no args passed)"
  query = query.query_params_with_all_defaults
}

control "query_params_with_defaults_and_partial_named_args" {
  title = "Control to test query param functionality with defaults(and some named args passed in query)"
  query = query.query_params_with_all_defaults
  args = {
    "p2" = "command_parameter_2"
  }
}

control "query_params_with_defaults_and_partial_positional_args" {
  title = "Control to test query param functionality with defaults(and some positional args passed in query)"
  query = query.query_params_with_all_defaults
  args = [  "command_parameter_1" ]
}

control "query_params_with_defaults_and_all_named_args" {
  title = "Control to test query param functionality with defaults(and all named args passed in query)"
  query = query.query_params_with_all_defaults
  args = {
    "p1" = "command_parameter_1"
    "p2" = "command_parameter_2"
    "p3" = "command_parameter_3"
  }
}

control "query_params_with_defaults_and_all_positional_args" {
  title = "Control to test query param functionality with defaults(and all positional args passed in query)"
  query = query.query_params_with_all_defaults
  args = [  "command_parameter_1", "command_parameter_2", "command_parameter_3" ]
}

control "query_params_with_no_defaults_and_no_args" {
  title = "Control to test query param functionality with no defaults(and no args passed)"
  query = query.query_params_with_no_defaults
}

control "query_params_with_no_defaults_with_named_args" {
  title = "Control to test query param functionality with no defaults(and args passed in query)"
  query = query.query_params_with_no_defaults
  args = {
    "p1" = "command_parameter_1"
    "p2" = "command_parameter_2"
    "p3" = "command_parameter_3"
  }
}

control "query_params_with_no_defaults_with_positional_args" {
  title = "Control to test query param functionality with no defaults(and positional args passed in query)"
  query = query.query_params_with_no_defaults
  args = [  "command_parameter_1", "command_parameter_2","command_parameter_3" ]
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

control "query_inline_sql_from_control_with_partial_named_args" {
  title = "Control to test the inline sql functionality within a control with defaults(and some named args passed in control)"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
        description = "p1"
        default = "default_parameter_1"
    }
    param "p2"{
        description = "p2"
        default = "default_parameter_2"
    }
    param "p3"{
        description = "p3"
        default = "default_parameter_3"
    }
    args = {
        "p1" = "command_parameter_1"
        "p3" = "command_parameter_3"
    }
  }

control "query_inline_sql_from_control_with_partial_positional_args" {
  title = "Control to test the inline sql functionality within a control with defaults(and some positional args passed in control)"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
        description = "p1"
        default = "default_parameter_1"
    }
    param "p2"{
        description = "p2"
        default = "default_parameter_2"
    }
    param "p3"{
        description = "p3"
        default = "default_parameter_3"
    }
    args = [  "command_parameter_1", "command_parameter_2" ]
  }

control "query_inline_sql_from_control_with_no_args" {
  title = "Control to test the inline sql functionality within a control with defaults(and no args passed in control)"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
        description = "p1"
        default = "default_parameter_1"
    }
    param "p2"{
        description = "p2"
        default = "default_parameter_2"
    }
    param "p3"{
        description = "p3"
        default = "default_parameter_3"
    }
  }

control "query_inline_sql_from_control_with_all_positional_args" {
  title = "Control to test the inline sql functionality within a control with defaults(and all positional args passed in control)"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
        description = "p1"
        default = "default_parameter_1"
    }
    param "p2"{
        description = "p2"
        default = "default_parameter_2"
    }
    param "p3"{
        description = "p3"
        default = "default_parameter_3"
    }
    args = [  "command_parameter_1", "command_parameter_2", "command_parameter_3" ]
  }

control "query_inline_sql_from_control_with_all_named_args" {
  title = "Control to test the inline sql functionality within a control with defaults(and all named args passed in control)"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
        description = "p1"
        default = "default_parameter_1"
    }
    param "p2"{
        description = "p2"
        default = "default_parameter_2"
    }
    param "p3"{
        description = "p3"
        default = "default_parameter_3"
    }
    args = {
        "p1" = "command_parameter_1"
        "p2" = "command_parameter_2"
        "p3" = "command_parameter_3"
    }
  }