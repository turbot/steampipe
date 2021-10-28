benchmark "check_search_path_benchmark" {
  title         = "Benchmark to test search path and search path prefix functionalities in steampipe check"
  children = [
    control.search_path_test_1,
    control.search_path_test_2,
    control.search_path_test_3,
    control.search_path_test_4,
    control.search_path_test_5,
    control.search_path_test_6
  ]
}

control "search_path_test_1" {
  title         = "Control to test search path prefix functionality when entered through CLI"
  description   = "Control to test search path prefix functionality when entered through CLI."
  sql           = query.search_path_1.sql
  severity      = "high"
}

control "search_path_test_2" {
  title         = "Control to test search path functionality when entered through CLI"
  description   = "Control to test search path functionality when entered through CLI."
  sql           = query.search_path_2.sql
  severity      = "high"
}

control "search_path_test_3" {
  title         = "Control to test search path and prefix functionality when entered through CLI"
  description   = "Control to test search path and prefix functionality when entered through CLI."
  sql           = query.search_path_1.sql
  severity      = "high"
}

control "search_path_test_4" {
  title         = "Control to test search path prefix functionality when entered through control"
  description   = "Control to test search path prefix functionality when entered through control."
  sql           = query.search_path_1.sql
  search_path_prefix   = "aws"
  severity      = "high"
}

control "search_path_test_5" {
  title         = "Control to test search path functionality when entered through control"
  description   = "Control to test search path functionality when entered through control."
  sql           = query.search_path_2.sql
  search_path   = "chaos,b,c"
  severity      = "high"
}

control "search_path_test_6" {
  title         = "Control to test search path and prefix functionality when entered through control"
  description   = "Control to test search path and prefix functionality when entered through control."
  sql           = query.search_path_1.sql
  search_path_prefix   = "aws"
  search_path   = "a,b,c"
  severity      = "high"
}