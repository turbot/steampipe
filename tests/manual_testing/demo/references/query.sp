variable "v1"{
    type = string
    default = "v1"
}

variable "v2"{
    type = string
    default = "v1"
}


query "q1"{
    title ="Q1"
    description = var.v1
    sql = "select 'ok' as status, 'foo' as resource, concat($1::text, $2::text, $3::text) as reason"
    param "p1"{
        description = "p1"
        default = var.v1
    }
    param "p2"{
        description = "p2"
        default = var.v1
    }
    param "p3"{
        description = "p3"
        default = var.v2
    }
}

