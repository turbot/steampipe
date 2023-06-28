benchmark "check_all" {
  title = "Benchmark to test the steampipe check all functionality"
  children = [
    control.check_1,
    control.check_2
  ]
}

control "check_1" {
  title         = "Control to verify steampipe check all functionality 1"
  description   = "Control to verify steampipe check all functionality."
  query         = query.query_1
  severity      = "high"
}

control "check_2" {
  title         = "Control to verify steampipe check all functionality 2"
  description   = "Control to verify steampipe check all functionality."
  query         = query.query_2
  severity      = "critical"
}