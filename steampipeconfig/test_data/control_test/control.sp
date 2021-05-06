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
        control.c20,
        control.c4,
        control.c5,
    ]
}

control "c1"{
    title = "control 1"
    sql = "select 'r1' as resource, 'alarm' as status, 'Im alarmed' as reason"
}
control "c2"{
    title = "control 2"
    sql = "select 'r2' as resource, 'alarm' as status, 'Im alarmed' as reason"
}
control "c3"{
    title = "control 3"
    sql = "select 'r3' as resource, 'alarm' as status, 'Im alarmed' as reason"
}
control "c4"{
    title = "control 4"
    sql = "select 'r4' as resource, 'alarm' as status, 'Im alarmed' as reason"
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
