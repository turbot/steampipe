control "c1"{
    title ="C1"
    description = "THIS IS CONTROL 1"
    query = query.q1
}

control "c2"{
    title ="C2"
    description = "THIS IS CONTROL 2"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::jsonb, $3::text) as reason"
    param "foo"{
        description = "p1"
        default = "default_control "
    }
    param "bar"{
        description = "p2"
        default = "because_def "
    }
    param "bill"{
        description = "p3"
        default = "string"
    }
}

control "c3"{
    title ="C3"
    description = "THIS IS CONTROL 3"
    query = query.q1
    args = [  "control3____ ", "because FOO ______ " ]
}

