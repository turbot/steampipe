benchmark "cg_1"{
    children = [
        benchmark.cg_1_1,
        benchmark.cg_1_2
    ]
}
benchmark "cg_1_1"{
    children = [
        benchmark.cg_1_1_1,
        benchmark.cg_1_1_2,
        control.c3,
    ]
}
benchmark "cg_1_2"{
}
benchmark "cg_1_1_1"{
    children = [
        control.c1,
    ]
}
benchmark "cg_1_1_2"{
    children = [
        contro1.c0,
        control.c4,
        control.c5,
        local.s2
    ]
}

control "c1"{
    title = "control 1"
    sql = "select 'r1' as resource, 'alarm' as status, 'Im alarmed' as reason"
}
control "c2"{
    title = "control 2"
    sql = control.c3.sql
}
control "c3"{
    title = "control 3"
    sql = control.c4.sql
}
control "c4"{
    title = "control 4"
    sql = control.foo.sql
    severity = "terrible"
}
control "c5"{
    title = "control 5"
    sql = "select 'r5' as resource, 'alarm' as status, 'Im alarmed' as reason"
}
control "c6"{
    title = "control 6"
    sql = "select 'r6' as resource, 'alarm' as status, 'Im alarmed' as reason"
}
