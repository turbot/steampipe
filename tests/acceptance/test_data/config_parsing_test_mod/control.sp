benchmark "config_parsing_benchmark" {
  title = "Benchmark to verify that the options config is parsed and used, by checking the cache functionality"
  children = [
    control.cache_test_11,
    control.cache_test_12
  ]
}

control "cache_test_11" {
  title         = "Control to verify that the options config is parsed and used 1"
  description   = "Control to verify that the options config is parsed and used."
  query           = query.chaos6_query
  severity      = "high"
}

control "cache_test_12" {
  title         = "Control to verify that the options config is parsed and used 2"
  description   = "Control to verify that the options config is parsed and used."
  query           = query.chaos6_query
  severity      = "high"
}