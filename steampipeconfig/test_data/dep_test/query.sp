query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select 1"
}

query "cg"{
    sql = "select resource_name from steampipe_control where parent='control_group.cg_1_1'"
}

query "q2"{
    title ="Q1"
    description = "THIS IS QUERY 2"
    sql = query.q3.sql
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
    sql = query.q6.sql
}

query "q6"{
    title ="Q1"
    description = "THIS IS QUERY 6"
    sql ="OK THIS WILL FAIL"
}
