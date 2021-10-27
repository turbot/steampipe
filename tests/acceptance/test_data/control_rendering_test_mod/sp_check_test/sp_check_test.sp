benchmark "control_summary_and_check_benchmark" {
  title = "Benchmark to test the check summary output and different output & export formats in steampipe"
  children = [
    control.sample_control_mixed_results_1,
    control.sample_control_mixed_results_2
  ]
}

control "sample_control_mixed_results_1" {
  title         = "Sample control with all possible statuses(severity=high)"
  description   = "Sample control that returns 10 OK, 5 ALARM, 2 ERROR, 1 SKIP and 3 INFO"
  query         = query.generic_query
  severity      = "high"
  args = {
    "number_of_ok" = 10
    "number_of_alarm" = 5
    "number_of_error" = 2
    "number_of_skip" = 1
    "number_of_info" = 3
  }
}

control "sample_control_mixed_results_2" {
  title         = "Sample control with all possible statuses(severity=critical)"
  description   = "Sample control that returns 5 OK, 5 ALARM"
  query         = query.generic_query
  severity      = "critical"
  args = {
    "number_of_ok" = 5
    "number_of_alarm" = 5
  }
}