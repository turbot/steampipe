
control "query_params_with_defaults_and_partial_named_args" {
    title = "Control to test query param functionality with defaults(and some named args passed in query)"
    query = query.query_params_with_no_defaults
    args = {
        "p1" = "command_parameter_1"

    }
}

query "query_params_with_no_defaults"{
    description = "query 1 - 3 params with no defaults"
    sql = "select $1::text[]"
    param "p1"{
        description = "First parameter"
        default = ["c","d"]
    }

}