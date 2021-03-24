mod "m1"{
  title = "M1"
  description = "THIS IS M1"
  version= "0.0.0"

  mod_depends{
    name = "github.com/turbot/m2"
    version = "0.0.0"
  }

  query "q1" {
    title ="Q1"
    description = "THIS IS QUERY 1"
    sql = "select 1"
  }
}