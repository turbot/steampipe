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
    query = "select 'pass' as result"
    parent = "control_group.cg_1_1_1"
}
control "c2"{
    query = "select 'pass' as result"
    parent = "control_group.cg_1_1_2"
}
control "c3"{
    query = "select 'pass' as result"
    parent = "control_group.cg_1_1"
}
control "c4"{
    query = "select 'pass' as result"
    parent = "control_group.cg_1_1_2"
}
control "c5"{
    query = "select 'pass' as result"
    parent = "control_group.cg_1_1_2"
}
control "c6"{
    query = "select 'FAIL' as result"
    // no parent - under mod
}
