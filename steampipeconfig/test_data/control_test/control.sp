control_group "cg_1"{
}
control_group "cg_1_1"{
    parent = control_group.cg_1
}
control_group "cg_1_2"{
    parent = control_group.cg_1
}
control_group "cg_1_1_1"{
    parent = control_group.cg_1_1
}
control_group "cg_1_1_2"{
    parent = control_group.cg_1_1
}

control "c1"{
    title = "control 1"
    sql = "select 'r1' as resource, 'alarm' as status, 'Im alarmed' as reason"
    parent = control_group.cg_1_1_1
}
control "c2"{
    title = "control 2"
    sql = "select 'r2' as resource, 'alarm' as status, 'Im alarmed' as reason"
    parent = control_group.cg_1_1_2
}
control "c3"{
    title = "control 3"
    sql = "select 'r3' as resource, 'alarm' as status, 'Im alarmed' as reason"
    parent = control_group.cg_1_1
}
control "c4"{
    title = "control 4"
    sql = "select 'r4' as resource, 'alarm' as status, 'Im alarmed' as reason"
    severity = "terrible"
    parent = control_group.cg_1_1_2
}
control "c5"{
    title = "control 5"
    sql = "select 'r5' as resource, 'alarm' as status, 'Im alarmed' as reason"
    parent = control_group.cg_1_1_2
}
control "c6"{
    title = "control 6"
    sql = "select 'r6' as resource, 'alarm' as status, 'Im alarmed' as reason"
    // no parent - under mod
}
