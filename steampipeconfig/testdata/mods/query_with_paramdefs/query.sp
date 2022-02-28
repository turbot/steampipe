query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select $1"

    param "p1"{
        description = "desc"
        default = "I am default"
    }

}
