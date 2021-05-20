locals {
    ll35858 = "testing"
    ll4 = "select 4 as foo"
}


query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select 1"
}

query "cg"{
    sql = "select resource_name from steampipe_control where parent='benchmark.cg_1_1'"
}

query "q2"{
    title ="Q1"
    description = "THIS IS QUERY 2"
    sql = query.q1.name
}

query "q3"{
    title ="Q1"
    description = "THIS IS QUERY 3"
    sql = query.q4.sql
}
query "q4"{
    title ="Q1"
    description = "THIS IS QUERY 4"
    sql = query.q5.sql
}

query "q5"{
    title ="Q1"
    description = "THIS IS QUERY 5"
    sql = local.ll4
}

query "q6"{
    title ="Q1"
    description = "THIS IS QUERY 6"
    sql = local.q2
}
