
control "c1"{
    title ="C1"
    description = "THIS IS CONTROL 1"
    query = query.q1
}

control "c2"{
    title ="C2"
    description = "THIS IS CONTROL 2"
    query = query.q1
    args = {
        "p1" = "control2 "
        "p3" = "a reason"
    }
}

control "c3"{
    title ="C3"
    description = "THIS IS CONTROL 3"
    query = query.q1
    args = [  "control3____ ", "because FOO ______ " ]
}

control "c4"{
    title ="C4"
    description = "THIS IS CONTROL 4"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = "c_default_control "
    }
    param "p2"{
        description = "p2"
        default = "c_because_def "
    }

    param "p3"{
        description = "p3"
        default = "c_string"
    }
}

control "c5"{
    title ="C5"
    description = "THIS IS CONTROL 5"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = "c_default_control "
    }
    param "p2"{
        description = "p2"
        default = "c_because_def "
    }
    param "p3"{
        description = "p3"
        default = "c_string"
    }
    args = [  "control5____ ", "because FOO_c5 ______ " ]
}
control "c5_this_is_a_very_long_name_no_even_longer_than_that_really_really_long_1"{
    title ="C5"
    description = "THIS IS CONTROL 5"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = "c_default_control "
    }
    param "p2"{
        description = "p2"
        default = "c_because_def "
    }
    param "p3"{
        description = "p3"
        default = "c_string"
    }
}

control "c5_this_is_a_very_long_name_no_even_longer_than_that_really_really_long_2"{
    title ="C5"
    description = "THIS IS CONTROL 5"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = "c_default_control "
    }
    param "p2"{
        description = "p2"
        default = "c_because_def "
    }
    param "p3"{
        description = "p3"
        default = "c_string"
    }
}


control "control_with_param_defauls_and_args"{
    title ="C5"
    description = "THIS IS CONTROL 5"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = "c_default_control "
    }
    param "p2"{
        description = "p2"
        default = "c_because_def "
    }
    param "p3"{
        description = "p3"
        default = "c_string"
    }
//    args  = {
//        p1 = "arg_control"
//    }
}
