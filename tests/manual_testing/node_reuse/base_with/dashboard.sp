
dashboard "my_dashboard_2" {
  table {
    base = table.t1
  }
}


table "t1"{
  sql = "select $1 as c1"
  param "p1" {
    default = "foo2"
  }
}
