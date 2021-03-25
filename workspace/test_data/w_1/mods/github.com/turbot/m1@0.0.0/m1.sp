mod "m1"{
  title = "M1"
  description = "THIS IS M1"
  version= "0.0.0"

  mod_depends{
    name = "github.com/turbot/m3"
    version = "0.0.0"
    alias = "m3"
  }

  query "q1" {
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select 1"
  }
}