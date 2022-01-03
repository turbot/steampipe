benchmark "nist_csf_de" {
  title       = "Detect (DE)"
  description = "Develop and implement appropriate activities to identify the occurrence of a cybersecurity event."
  children = [
    benchmark.nist_csf_de_ae,
    benchmark.nist_csf_de_cm,
    benchmark.nist_csf_de_dp
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_ae" {
  title       = "Anomalies and Events (DE.AE)"
  description = "Anomalous activity is detected and the potential impact of events is understood."
  children = [
    benchmark.nist_csf_de_ae_1,
    benchmark.nist_csf_de_ae_2,
    benchmark.nist_csf_de_ae_3,
    benchmark.nist_csf_de_ae_4,
    benchmark.nist_csf_de_ae_5
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_ae_1" {
  title       = "DE.AE-1"
  description = "A baseline of network operations and expected data flows for users and systems is established and managed."
  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.elb_application_classic_lb_logging_enabled,
    control.lambda_function_concurrent_execution_limit_configured,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.vpc_default_security_group_restricts_all_traffic,
    control.vpc_flow_logs_enabled,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_security_group_restrict_ingress_ssh_all,
    control.vpc_security_group_restrict_ingress_tcp_udp_all
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_ae_2" {
  title       = "DE.AE-2"
  description = "Detected events are analyzed to understand attack targets and methods."
  children = [
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_ae_3" {
  title       = "DE.AE-3"
  description = "Event data are collected and correlated from multiple sources and sensors."
  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.elb_application_classic_lb_logging_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.vpc_flow_logs_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_ae_4" {
  title       = "DE.AE-4"
  description = "Impact of events is determined."
  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.elb_application_classic_lb_logging_enabled,
    control.guardduty_enabled,
    control.guardduty_finding_archived,
    control.s3_bucket_logging_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_ae_5" {
  title       = "DE.AE-5"
  description = "Incident alert thresholds are established."

  children = [
    control.cloudwatch_alarm_action_enabled,
    control.codebuild_project_source_repo_oauth_configured,
    control.lambda_function_dead_letter_queue_configured
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm" {
  title       = "Security Continuous Monitoring (DE.CM)"
  description = "The information system and assets are monitored to identify cybersecurity events and verify the effectiveness of protective measures."
  children = [
    benchmark.nist_csf_de_cm_1,
    benchmark.nist_csf_de_cm_2,
    benchmark.nist_csf_de_cm_3,
    benchmark.nist_csf_de_cm_4,
    benchmark.nist_csf_de_cm_5,
    benchmark.nist_csf_de_cm_6,
    benchmark.nist_csf_de_cm_7
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_1" {
  title       = "DE.CM-1"
  description = "The network is monitored to detect potential cybersecurity events."

  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.elb_application_classic_lb_logging_enabled,
    control.guardduty_enabled,
    control.s3_bucket_logging_enabled,
    control.securityhub_enabled,
    control.vpc_flow_logs_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_2" {
  title       = "DE.CM-2"
  description = "The physical environment is monitored to detect potential cybersecurity events."

  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.config_enabled_all_regions,
    control.ec2_instance_detailed_monitoring_enabled,
    control.guardduty_enabled,
    control.log_metric_filter_bucket_policy,
    control.log_metric_filter_cloudtrail_configuration,
    control.log_metric_filter_config_configuration,
    control.log_metric_filter_console_authentication_failure,
    control.log_metric_filter_console_login_mfa,
    control.log_metric_filter_disable_or_delete_cmk,
    control.log_metric_filter_iam_policy,
    control.log_metric_filter_network_acl,
    control.log_metric_filter_network_gateway,
    control.log_metric_filter_root_login,
    control.log_metric_filter_route_table,
    control.log_metric_filter_security_group,
    control.log_metric_filter_unauthorized_api,
    control.log_metric_filter_vpc,
    control.rds_db_instance_and_cluster_enhanced_monitoring_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_3" {
  title       = "DE.CM-3"
  description = "Personnel activity is monitored to detect potential cybersecurity events."

  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.guardduty_enabled,
    control.s3_bucket_logging_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_4" {
  title       = "DE.CM-4"
  description = "Malicious code is detected."

  children = [
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_5" {
  title       = "DE.CM-5"
  description = "Unauthorized mobile code is detected."

  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.ec2_instance_detailed_monitoring_enabled,
    control.elb_application_lb_waf_enabled,
    control.guardduty_enabled,
    control.guardduty_finding_archived,
    control.securityhub_enabled,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_6" {
  title       = "DE.CM-6"
  description = "External service provider activity is monitored to detect potential cybersecurity events."

  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.guardduty_enabled,
    control.s3_bucket_logging_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_cm_7" {
  title       = "DE.CM-7"
  description = "Monitoring for unauthorized personnel, connections, devices, and software is performed."

  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.elb_application_classic_lb_logging_enabled,
    control.guardduty_enabled,
    control.s3_bucket_logging_enabled,
    control.securityhub_enabled,
    control.vpc_flow_logs_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_dp" {
  title       = "Detection Processes (DE.DP)"
  description = "Detection processes and procedures are maintained and tested to ensure awareness of anomalous events."
  children = [
    benchmark.nist_csf_de_dp_4,
    benchmark.nist_csf_de_dp_5
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_dp_4" {
  title       = "DE.DP-4"
  description = "Event detection information is communicated."

  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.ec2_instance_detailed_monitoring_enabled,
    control.elb_application_lb_waf_enabled,
    control.guardduty_enabled,
    control.guardduty_finding_archived,
    control.securityhub_enabled,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_de_dp_5" {
  title       = "DE.DP-5"
  description = "Detection processes are continuously improved."

  children = [
    control.ec2_instance_detailed_monitoring_enabled
  ]

  tags = local.nist_csf_common_tags
}
