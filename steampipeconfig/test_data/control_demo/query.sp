query "q1"{
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select 1"
}

query "cg"{
    sql = "select resource_name from steampipe_controls where parent='control_group.cg_1_1'"
}

