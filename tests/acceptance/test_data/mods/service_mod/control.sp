benchmark "check_all" {
  title = "Benchmark to test the steampipe service stability"
  children = [
    control.check_1,
    control.check_2
  ]
}

control "check_1" {
  title         = "Control 1"
  query         = query.query_1
  severity      = "high"
}

control "check_2" {
  title         = "Control 2"
  query         = query.query_2
  severity      = "critical"
}