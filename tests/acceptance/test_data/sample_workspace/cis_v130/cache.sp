benchmark "check_cache_benchmark" {
  title         = "Benchmark to test the cache functionality in steampipe"
  children = [
    control.cache_test_1
  ]
}

control "cache_test_1" {
  title         = "Control to test cache functionality"
  description   = "Control to test cache functionality in steampipe."
  sql           = query.check_cache.sql
  severity      = "high"
}