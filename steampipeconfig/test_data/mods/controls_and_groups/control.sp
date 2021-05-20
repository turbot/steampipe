benchmark "cg_1"{
}
benchmark "cg_1_1"{
    parent = "benchmark.cg_1"
}
benchmark "cg_1_2"{
    parent = "benchmark.cg_1"
}
benchmark "cg_1_1_1"{
    parent = "benchmark.cg_1_1"
}
benchmark "cg_1_1_2"{
    parent = "benchmark.cg_1_1"
}
control "c1"{
    sql = "select 'pass' as result"
    parent = "benchmark.cg_1_1_1"
}
control "c2"{
    sql = "select 'pass' as result"
    parent = "benchmark.cg_1_1_2"
}
control "c3"{
    sql = "select 'pass' as result"
    parent = "benchmark.cg_1_1"
}
control "c4"{
    sql = "select 'pass' as result"
    parent = "benchmark.cg_1_1_2"
}
control "c5"{
    sql = "select 'pass' as result"
    parent = "benchmark.cg_1_1_2"
}
control "c6"{
    sql = "select 'FAIL' as result"
    // no parent - under mod
}
