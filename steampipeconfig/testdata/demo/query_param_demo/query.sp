variable "v1"{
    type = string
    default = "from_var"
}


query "q1"{
    title ="Q1"
    description = "query 1 - 3 params all with defaults"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = var.v1
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


query "q2" {
    title       = "EC2 Instances xlarge and bigger"
    sql = "select 'ok' as status, 'foo' as resource, $1::jsonb->'a' as reason"
    param "p1"{
        description = "p1"
    }
}

query "q3" {
    sql = "select * from chaos_all_column_types where string_column like any($1)"
    param "p1"{
        default = ["stringValuesomething-13","stringValuesomething-7"]
    }
}
