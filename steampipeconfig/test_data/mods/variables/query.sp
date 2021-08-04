variable "v1"{
    type = string
    default = "select 'default'"
}

variable "v2"{
    type = list(string)
    default = ["select 'default1'", "select 'default2'", "select 'default3'"]
}

variable "v3" {
    type = list(object({
        internal = number
        external = number
        query = string
    }))
    default = [
        {
            internal = 8300
            external = 8300
            query = "select 'default4'"
        }
    ]
}

query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = variable.v1
}

query "q2"{
    title ="Q2"
    description = "THIS IS QUERY 2"
    sql = variable.v2[0]
}

query "q3"{
    title ="Q3"
    description = query.q1.description
    sql = variable.v3[0].query
}
