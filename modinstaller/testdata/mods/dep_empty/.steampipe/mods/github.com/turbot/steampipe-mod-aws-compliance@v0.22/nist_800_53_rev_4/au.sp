benchmark "nist_800_53_rev_4_au" {
  title       = "Audit and Accountability (AU)"
  description = "The AU control family consists of security controls related to an organization’s audit capabilities. This includes audit policies and procedures, audit logging, audit report generation, and protection of audit information."
  children = [
    benchmark.nist_800_53_rev_4_au_2,
    benchmark.nist_800_53_rev_4_au_3,
    benchmark.nist_800_53_rev_4_au_6,
    benchmark.nist_800_53_rev_4_au_7,
    benchmark.nist_800_53_rev_4_au_9,
    benchmark.nist_800_53_rev_4_au_11,
    benchmark.nist_800_53_rev_4_au_12
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_2" {
  title       = "Event Logging (AU-2)"
  description = "Automate security audit function with other organizational entities. Enable mutual support of audit of auditable events."
  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.elb_application_classic_lb_logging_enabled,
    control.rds_db_instance_logging_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.vpc_flow_logs_enabled,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_3" {
  title       = "Content of Audit Records (AU-3)"
  description = "The information system generates audit records containing information that establishes what type of event occurred, when the event occurred, where the event occurred, the source of the event, the outcome of the event, and the identity of any individuals or subjects associated with the event."
  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.elb_application_classic_lb_logging_enabled,
    control.rds_db_instance_logging_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.vpc_flow_logs_enabled,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_6" {
  title       = "Audit Review, Analysis And Reporting (AU-6)"
  description = "Integrate audit review, analysis, and reporting with processes for investigation and response to suspicious activities."
  children = [
    benchmark.nist_800_53_rev_4_au_6_1,
    benchmark.nist_800_53_rev_4_au_6_3
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_6_1" {
  title       = "AU-6(1) Process Integration"
  description = "The organization employs automated mechanisms to integrate audit review, analysis,and reporting processes to support organizational processes for investigation and response to suspicious activities."
  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_6_3" {
  title       = "AU-6(3) Correlate Audit Repositories"
  description = "The organization analyzes and correlates audit records across different repositories to gain organization-wide situational awareness."
  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_7" {
  title       = "Audit Reduction And Report Generation (AU-7)"
  description = "Support for real-time audit review, analysis, and reporting requirements without altering original audit records."
  children = [
    benchmark.nist_800_53_rev_4_au_7_1
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_7_1" {
  title       = "AU-7(1) Automatic Processing"
  description = "The information system provides the capability to process audit records for events of interest based on [Assignment: organization-defined audit fields within audit records]."
  children = [
    control.cloudwatch_alarm_action_enabled,
    control.cloudtrail_trail_integrated_with_logs
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_9" {
  title       = "Protection of Audit Information (AU-9)"
  description = "The information system protects audit information and audit tools from unauthorized access, modification, and deletion."
  children = [
    control.cloudtrail_trail_logs_encrypted_with_kms_cmk,
    control.log_group_encryption_at_rest_enabled,
    benchmark.nist_800_53_rev_4_au_9_2
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_9_2" {
  title       = "AU-9(2) Audit Backup On Separate Physical Systems / Components"
  description = "The information system backs up audit records [Assignment: organization-defined frequency] onto a physically different system or system component than the system or component being audited."
  children = [
    control.s3_bucket_cross_region_replication_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_11" {
  title       = "Audit Record Retention (AU-11)"
  description = "The organization retains audit records for [Assignment: organization-defined time period consistent with records retention policy] to provide support for after-the-fact investigations of security incidents and to meet regulatory and organizational information retention requirements."
  children = [
    control.cloudwatch_log_group_retention_period_365
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_au_12" {
  title       = "Audit Generation (AU-12)"
  description = "Audit events defined in AU-2. Allow trusted personnel to select which events to audit. Generate audit records for events."
  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.elb_application_classic_lb_logging_enabled,
    control.rds_db_instance_logging_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.vpc_flow_logs_enabled,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
