control "c1"{
  sql = "select 'pass' as result"
}

report "r1"{
  control {
    base = control.c1
  }
}