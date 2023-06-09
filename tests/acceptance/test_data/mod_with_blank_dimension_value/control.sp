control "check_1" {
  title         = "Control to verify steampipe check all functionality 1"
  description   = "Control to verify steampipe check all functionality."
  query         = query.control_with_blank_dimension
  severity      = "high"
}
