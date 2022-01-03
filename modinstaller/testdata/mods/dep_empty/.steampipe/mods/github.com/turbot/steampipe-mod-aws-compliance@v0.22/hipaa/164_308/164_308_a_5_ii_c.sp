benchmark "hipaa_164_308_a_5_ii_c" {
  title       = "164.308(a)(5)(ii)(C) Log-in monitoring"
  description = "Procedures for monitoring log-in attempts and reporting discrepancies."
  children = [
    control.guardduty_enabled,
    control.log_metric_filter_console_authentication_failure,
    control.securityhub_enabled
  ]

  tags = merge(local.hipaa_164_308_common_tags, {
    hipaa_item_id = "164_308_a_5_ii_c"
  })
}