benchmark "control_summary_benchmark" {
  title = "Benchmark to test the check summary output in steampipe"
  children = [
    control.sample_control_ok,
    control.sample_control_alarm,
    control.sample_control_info,
    control.sample_control_error,
    control.sample_control_skip
  ]
}

control "sample_control_ok" {
  title         = "Sample control OK"
  description   = "A sample control that returns OK"
  sql           = query.ok.sql
  severity      = "high"
}

control "sample_control_alarm" {
  title         = "Sample control ALARM"
  description   = "A sample control that returns ALARM"
  sql           = query.alarm.sql
  severity      = "critical"
}

control "sample_control_info" {
  title         = "Sample control INFO"
  description   = "A sample control that returns INFO"
  sql           = query.info.sql
  severity      = "high"
}

control "sample_control_error" {
  title         = "Sample control ERROR"
  description   = "A sample control that returns ERROR"
  sql           = query.error.sql
  severity      = "critical"
}

control "sample_control_skip" {
  title         = "Sample control SKIP"
  description   = "A sample control that returns SKIP"
  sql           = query.skip.sql
  severity      = "high"
}