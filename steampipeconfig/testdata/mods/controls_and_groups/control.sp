benchmark "cg_1"{
    children = [benchmark.cg_1_1, benchmark.cg_1_2 ]
}

benchmark "cg_1_1"{
    children = [benchmark.cg_1_1_1, benchmark.cg_1_1_2]
}

benchmark "cg_1_2"{
}

benchmark "cg_1_1_1"{
    children = [control.c1]
}

benchmark "cg_1_1_2"{
    children = [control.c2, control.c4, control.c5]
}

control "c1"{
    sql = "select 'pass' as result"
}

control "c2"{
    sql = "select 'pass' as result"
}

control "c3"{
    sql = "select 'pass' as result"
}

control "c4"{
    sql = "select 'pass' as result"
}

control "c5"{
    sql = "select 'pass' as result"
}

control "c6"{
    sql = "select 'fail' as result"
}
