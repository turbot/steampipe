query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select 1"
}

query "cg"{
    sql = "select resource_name from steampipe_control where parent='benchmark.cg_1_1'"
}

