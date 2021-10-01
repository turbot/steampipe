benchmark "control_summary_benchmark" {
  title = "Benchmark to test the check summary output in steampipe"
  children = [
    control.sample_control_1,
    control.sample_control_2,
    control.sample_control_3,
    control.sample_control_4,
    control.sample_control_5
  ]
}

control "sample_control_1" {
  title         = "Sample control 1"
  description   = "A sample control"
  sql           = query.static_query.sql
  severity      = "high"
}

control "sample_control_2" {
  title         = "Sample control 2"
  description   = "A sample control"
  sql           = query.static_query.sql
  severity      = "critical"
}

control "sample_control_3" {
  title         = "Sample control 3"
  description   = "A sample control"
  sql           = query.static_query.sql
  severity      = "high"
}

control "sample_control_4" {
  title         = "Sample control 4"
  description   = "A sample control that returns ERROR"
  sql           = query.static_query.sql
  severity      = "critical"
}

control "sample_control_5" {
  title         = "Sample control 5"
  description   = "A sample control"
  sql           = query.static_query.sql
  severity      = "high"
}