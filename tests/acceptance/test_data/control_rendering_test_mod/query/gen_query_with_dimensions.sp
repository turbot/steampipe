query "generic_query_with_dimensions" {
  description = "parameterized query to simulate control results, with rows conataining all possible statuses(with extra dimensions)"
  sql = query.gen_query_with_dimensions.sql
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
