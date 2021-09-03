variable "query1"{
    type = string
    description = "string variable with a default"
    default = "select 'var.query1'"
}
variable "column" {
    description = "string variable with no default"
    type=string
}

variable "regions"{
    type = list(string)
    description = "string array variable with default"
    default = ["eu-west2", "us-east1"]
}

variable "queries" {
    type = list(object({
        query = string
        metadata = string

    }))
    description = "object array variable with default"
    default = [
        {
            metadata = "foo"
            query = "select * from aws_account"
        },
        {
            metadata = "bar"
            query = "select * from aws_iam_group"
        }
    ]
}


query "q1"{
    description = "use variable within a string"
    sql = "select ${var.column}"
}

query "q2"{
    title ="Q2"
    description = "accounts"
    sql = var.queries[0].query
}
query "q3"{
    title ="Q2"
    description = "groups"
    sql = var.queries[1].query
}