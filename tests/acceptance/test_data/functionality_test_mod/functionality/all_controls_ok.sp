benchmark "all_controls_ok" {
  title         = "All controls in OK, no ALARMS/ERORS"
  description   = "Benchmark to verify the exit code when no controls are in error/alarm"
  children      = [
    control.ok_1,
    control.ok_2
  ]
}

control "ok_1" {
  title         = "Control to verify the exit code when no controls are in error/alarm"
  description   = "Control to verify the exit code when no controls are in error/alarm"
  query         = query.query_1
  severity      = "high"
}

control "ok_2" {
  title         = "Control to verify the exit code when no controls are in error/alarm"
  description   = "Control to verify the exit code when no controls are in error/alarm"
  query         = query.query_1
  severity      = "high"
}

query "query_1"{
  title ="query_1"
  description = "Simple query 1"
  sql = "select 'ok' as status, 'steampipe' as resource, 'acceptance tests' as reason"
}