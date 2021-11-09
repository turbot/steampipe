benchmark "check_cache_same_columns_benchmark" {
  title         = "Benchmark to test the cache functionality in steampipe when querying same columns"
  children      = [
    control.same_columns_1,
    control.same_columns_2
  ]
}

control "same_columns_1" {
  title         = "Query same columns 1"
  description   = "Control to test cache functionality in steampipe when querying same columns."
  sql           = query.check_cache.sql
  severity      = "high"
}

control "same_columns_2" {
  title         = "Query same columns 2"
  description   = "Control to test cache functionality in steampipe when querying same columns."
  sql           = query.check_cache.sql
  severity      = "high"
}


benchmark "check_cache_subset_columns_benchmark" {
  title         = "Benchmark to test the cache functionality in steampipe when second query's columns is a subset of the first"
  children      = [
    control.subset_columns_1,
    control.subset_columns_2
  ]
}

control "subset_columns_1" {
  title         = "Query subset columns 1"
  description   = "Control to test cache functionality in steampipe when second query's columns is a subset of the first."
  sql           = query.check_cache.sql
  severity      = "high"
}

control "subset_columns_2" {
  title         = "Query subset columns 2"
  description   = "Control to test cache functionality in steampipe when second query's columns is a subset of the first."
  sql           = query.check_cache_subset.sql
  severity      = "high"
}


benchmark "check_cache_multiple_same_columns_benchmark" {
  title         = "Benchmark to test the cache functionality for multiple(4) queries with same columns"
  children      = [
    control.multiple_columns_1,
    control.multiple_columns_2,
    control.multiple_columns_3,
    control.multiple_columns_4
  ]
}

control "multiple_columns_1" {
  title         = "Query multiple columns 1"
  description   = "Control to test cache functionality in steampipe for multiple(4) queries with same columns."
  sql           = query.check_cache.sql
  severity      = "high"
}

control "multiple_columns_2" {
  title         = "Query multiple columns 2"
  description   = "Control to test cache functionality in steampipe for multiple(4) queries with same columns."
  sql           = query.check_cache.sql
  severity      = "high"
}

control "multiple_columns_3" {
  title         = "Query multiple columns 3"
  description   = "Control to test cache functionality in steampipe for multiple(4) queries with same columns."
  sql           = query.check_cache.sql
  severity      = "high"
}

control "multiple_columns_4" {
  title         = "Query multiple columns 4"
  description   = "Control to test cache functionality in steampipe for multiple(4) queries with same columns."
  sql           = query.check_cache.sql
  severity      = "high"
}


benchmark "check_cache_multiple_subset_columns_benchmark" {
  title         = "Benchmark to test the cache functionality in steampipe when multiple query's columns are a subset of the first"
  children      = [
    control.multiple_subset_columns_1,
    control.multiple_subset_columns_2,
    control.multiple_subset_columns_3,
    control.multiple_subset_columns_4
  ]
}

control "multiple_subset_columns_1" {
  title         = "Multiple query subset columns 1"
  description   = "Control to test cache functionality in steampipe when multiple query's columns are a subset of the first."
  sql           = query.check_cache.sql
  severity      = "high"
}

control "multiple_subset_columns_2" {
  title         = "Multiple query subset columns 2"
  description   = "Control to test cache functionality in steampipe when multiple query's columns are a subset of the first."
  sql           = query.check_cache_subset.sql
  severity      = "high"
}

control "multiple_subset_columns_3" {
  title         = "Multiple query subset columns 3"
  description   = "Control to test cache functionality in steampipe when multiple query's columns are a subset of the first."
  sql           = query.check_cache_subset.sql
  severity      = "high"
}

control "multiple_subset_columns_4" {
  title         = "Multiple query subset columns 4"
  description   = "Control to test cache functionality in steampipe when multiple query's columns are a subset of the first."
  sql           = query.check_cache_subset.sql
  severity      = "high"
}