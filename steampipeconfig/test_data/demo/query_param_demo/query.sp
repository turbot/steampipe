variable "v1"{
    type = string
    default = "from_var"
}

variable "regions"{
    type = list(string)
    description = "string array variable with default"
    default = ["eu-west2", "us-east1"]
}



query "q1"{
    title ="Q1"
    description = "query 1 - 3 params all with defaults"
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    params "p1"{
        description = "p1"
        default = "'default_control '"
    }
    params "p2"{
        description = "p2"
    }
    params "p3"{
        description = "p3"
        default = "'default_reason'"
    }
}
