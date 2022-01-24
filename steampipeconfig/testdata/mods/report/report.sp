benchmark "b1"{
  children = [ control.c1  ]
}
control "c1"{
  sql = query.manual_control.sql
}

report "r1"{
  benchmark {
    base = benchmark.b1
  }
}
report "r2"{
  control {
    base = control.c1
  }
}