control_group "cg_1"{
}
control_group "cg_1_1"{
    parent = "control_group.cg_1"
}
control_group "cg_1_2"{
    parent = "control_group.cg_1"
}
control_group "cg_1_1_1"{
    parent = "control_group.cg_1_1"
}
control_group "cg_1_1_2"{
    parent = "control_group.cg_1_1"
}
control "c1"{
    description = "control 1"
    query = "query.q1"
    parent = "control_group.cg_1_1_1"
}
control "c2"{
    description = "control 2"
    query = "select 'control 2' as control, 'pass' as result"
    parent = "control_group.cg_1_1_2"
}
control "c3"{
    description = "control 3"
    query = "select 'control 3' as control, 'pass' as result"
    parent = "control_group.cg_1_1"
}
control "c4"{
    description = "control 4"
    query = "select 'control 4' as control, 'pass' as result"
    parent = "control_group.cg_1_1_2"
}
control "c5"{
    description = "control 5"
    query = "select 'control 5' as control, 'pass' as result"
    parent = "control_group.cg_1_1_2"
}
control "c6"{
    description = "control 6"
    query = "select 'control 6' as control, 'FAIL' as result"
    // no parent - under mod
}
