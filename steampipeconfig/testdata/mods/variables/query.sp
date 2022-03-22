variable "account"{
    type = string
    description = "the account to use"
}

variable "reason"{
    type = string
    description = "reason for failure"
    default = "check failed"
}

variable "regions"{
    type = list(string)
    description = "the available regions"
    default = ["eu-west2", "us-east1"]
}

variable "v1"{
    type = string
    default = "select 'default'"
}

variable "v2"{
    type = list(string)
    default=["select 1"]

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

variable "v4"{
    type = string
    description="this is v4"
}

query "q1"{
    title ="Q1"
    description = var.reason
    sql = "select ${var.regions[0]}"
}

query "q2"{
    title ="Q2"
    description = "THIS IS QUERY 2"
    sql = var.v2[0]
}

query "q3"{
    title ="Q3"
    description = query.q1.description
    sql = var.v3[0].query
}
