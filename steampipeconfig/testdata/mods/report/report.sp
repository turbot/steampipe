control "c1"{
  sql = query.manual_control.sql
}

report "r1"{
  control {
    base = control.c1
  }
}