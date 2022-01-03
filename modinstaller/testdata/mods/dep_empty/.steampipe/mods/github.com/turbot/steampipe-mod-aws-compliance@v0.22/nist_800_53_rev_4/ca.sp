benchmark "nist_800_53_rev_4_ca" {
  title       = "Security Assessment And Authorization (CA)"
  description = "The Security Assessment and Authorization control family includes controls that supplement the execution of security assessments, authorizations, continuous monitoring, plan of actions and milestones, and system interconnections."
  children = [
    benchmark.nist_800_53_rev_4_ca_7
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ca_7" {
  title       = "Continuous Monitoring (CA-7)"
  description = "Continuously monitor configuration management processes. Determine security impact, environment and operational risks."
  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.ec2_instance_detailed_monitoring_enabled,
    control.guardduty_enabled,
    control.rds_db_instance_and_cluster_enhanced_monitoring_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
