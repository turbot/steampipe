
dashboard "param_ref" {

  table {
    base = table.t1

  }
}

table "t1"{

  with "w1" {
    sql = "select 'foo'"
  }

  sql = "select $1 as c1"
  param "p1" {
    default = with.w1.rows[0]
  }
}
