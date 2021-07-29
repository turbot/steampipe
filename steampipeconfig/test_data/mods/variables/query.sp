variable "v1"{
    type = string
    default = "select 1"
}
variable "v2"{
    type = list(string)
    default = ["select 1", "select 2", "select 3"]
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
            query = "select 3"
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
    sql = variable.v4
}
