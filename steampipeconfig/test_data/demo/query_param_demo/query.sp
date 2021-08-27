variable "v1"{
    type = string
    default = "from_var"
}


query "q1"{
    title ="Q1"
    description = "query 1 - 3 params all with defaults"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    params "p1"{
        description = "p1"
        default = "default_control "
    }
    params "p2"{
        description = "p2"
        default = ["default_because_${var.v1} ", 1]
    }
    params "p3"{
        description = "p3"
        default = 100
    }
}


query "q2" {
    title       = "EC2 Instances xlarge and bigger"
    sql = "select 'ok' as status, 'foo' as resource, $1::json->1 as reason"
}