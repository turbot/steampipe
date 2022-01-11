
benchmark "cg_1_1_2"{
    children = [control.c2, control.c4, control.c5]
}

control "c1"{
    sql = "select 'pass' as result"
}