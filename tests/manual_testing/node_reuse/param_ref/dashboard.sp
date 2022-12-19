
dashboard "param_ref" {

  table {
    base = table.t1

  }
}


table "t1"{
  param "dash" {
    default = "foo"
  }

  sql = "select $1 as c1"
  args = [ param.dash]
}
