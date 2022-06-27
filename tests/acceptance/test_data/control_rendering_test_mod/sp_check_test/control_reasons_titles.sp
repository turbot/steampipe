benchmark "control_reasons_and_titles_benchmark" {
  title = "Benchmark to test control reasons and titles(of different possible lengths) in steampipe"
  children = [
    control.control_long_title,
    control.control_short_title,
    control.control_unicode_title,
    control.control_long_short_unicode_reasons
  ]
}

control "control_long_title" {
  title         = "Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title Control with long title"
  description   = "Sample control with a very long title."
  query         = query.generic_query
  severity      = "high"
  args = {
    "number_of_ok" = 3
    "number_of_alarm" = 2
  }
}

control "control_short_title" {
  title         = "Control short title"
  description   = "Sample control with a very short title."
  query         = query.generic_query
  severity      = "critical"
  args = {
    "number_of_ok" = 3
    "number_of_alarm" = 2
  }
}

control "control_unicode_title" {
  title         = "Control unicode title ❌"
  description   = "Sample control with a title that contains unicode characters."
  query         = query.generic_query
  severity      = "critical"
  args = {
    "number_of_alarm" = 1
  }
}

control "control_long_short_unicode_reasons" {
  title         = "Control with long, short and unicode reasons"
  description   = "Sample control with few resources, one with a very short reason and the other with a very long reason, and one with an unicode character in the reason."
  sql           = query.long_short_unicode_reasons.sql
  severity      = "critical"
}
