query "generic_query" {
  description = "parameterized query to simulate control results, with rows conataining all possible statuses"
  sql = query.gen_query.sql
  param "number_of_ok" {
    description = "Number of resources in OK"
    default = 0
  }
  param "number_of_alarm" {
    description = "Number of resources in ALARM"
    default = 0
  }
  param "number_of_error" {
    description = "Number of resources in ERROR"
    default = 0
  }
  param "number_of_skip" {
    description = "Number of resources in SKIP"
    default = 0
  }
  param "number_of_info" {
    description = "Number of resources in INFO"
    default = 0
  }
}
