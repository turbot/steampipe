benchmark "nist_800_53_rev_4_si" {
  title       = "System and Information integrity (SI)"
  description = "The SI control family correlates to controls that protect system and information integrity. These include flaw remediation, malicious code protection, information system monitoring, security alerts, software and firmware integrity, and spam protection."
  children = [
    benchmark.nist_800_53_rev_4_si_2,
    benchmark.nist_800_53_rev_4_si_4,
    benchmark.nist_800_53_rev_4_si_7,
    benchmark.nist_800_53_rev_4_si_12
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_2" {
  title       = "Flaw Remediation (SI-2)"
  description = "The organization: a.Identifies, reports, and corrects information system flaws; b.Tests software and firmware updates related to flaw remediation for effectiveness and potential side effects before installation; c.Installs security-relevant software and firmware updates within [Assignment: organization-defined time period] of the release of the updates; and d.Incorporates flaw remediation into the organizational configuration management process."
  children = [
    benchmark.nist_800_53_rev_4_si_2_2,
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_2_2" {
  title       = "SI-2(2) Automates Flaw Remediation Status"
  description = "The organization employs automated mechanisms to determine the state of information system components with regard to flaw remediation."
  children = [
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant,
    control.ssm_managed_instance_compliance_patch_compliant,
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_4" {
  title       = "Information System Monitoring (SI-4)"
  description = "The organization: a.Monitors the information system to detect: 1. Attacks and indicators of potential attacks in accordance with [Assignment: organization-defined monitoring objectives]; and 2.Unauthorized local, network, and remote connections; b. Identifies unauthorized use of the information system through [Assignment: organization-defined techniques and methods]; c. Deploys monitoring devices: 1. Strategically within the information system to collect organization-determined essential information; and 2. At ad hoc locations within the system to track specific types of transactions of interest to the organization; d. Protects information obtained from intrusion-monitoring tools from unauthorized access, modification, and deletion; e. Heightens the level of information system monitoring activity whenever there is an indication of increased risk to organizational operations and assets, individuals, other organizations, or the Nation based on law enforcement information, intelligence information, or other credible sources of information; f. Obtains legal opinion with regard to information system monitoring activities in accordance with applicable federal laws, Executive Orders, directives, policies, or regulations; and g. Provides [Assignment: organization-defined information system monitoring information] to [Assignment: organization-defined personnel or roles] [Selection (one or more): as needed; [Assignment: organization-defined frequency]]."
  children = [
    benchmark.nist_800_53_rev_4_si_4_1,
    benchmark.nist_800_53_rev_4_si_4_2,
    benchmark.nist_800_53_rev_4_si_4_4,
    benchmark.nist_800_53_rev_4_si_4_5,
    benchmark.nist_800_53_rev_4_si_4_16,
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.ec2_instance_detailed_monitoring_enabled,
    control.elb_application_lb_waf_enabled,
    control.guardduty_enabled,
    control.guardduty_finding_archived,
    control.securityhub_enabled,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_4_1" {
  title       = "SI-4(1) System-Wide Intrusion Detection System"
  description = "The organization connects and configures individual intrusion detection tools into an information system-wide intrusion detection system."
  children = [
    control.guardduty_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_4_2" {
  title       = "SI-4(2) Automated Tools For Real-Time Analysis"
  description = "The organization employs automated tools to support near real-time analysis of events."
  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.ec2_instance_detailed_monitoring_enabled,
    control.guardduty_enabled,
    control.securityhub_enabled,
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_4_4" {
  title       = "SI-4(4) Inbound and Outbound Communications Traffic"
  description = "The information system monitors inbound and outbound communications traffic continuously for unusual or unauthorized activities or conditions."
  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_4_5" {
  title       = "SI-4(5) System-Generated Alerts"
  description = "The information system alerts organization-defined personnel or roles when the following indications of compromise or potential compromise occur: [Assignment: organization-defined compromise indicators]."
  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_4_16" {
  title       = "SI-4(16) Correlate Monitoring Information"
  description = "The organization correlates information from monitoring tools employed throughout the information system."
  children = [
    control.guardduty_enabled,
    control.securityhub_enabled,
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_7" {
  title       = "Software, Firmware, and Information Integrity (SI-7)"
  description = "The organization employs integrity verification tools to detect unauthorized changes to [Assignment: organization-defined software, firmware, and information]."
  children = [
    control.cloudtrail_trail_validation_enabled,
    benchmark.nist_800_53_rev_4_si_7_1
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_7_1" {
  title       = "SI-7(1) Integrity Checks"
  description = "The information system performs an integrity check of security relevant events at least monthly."
  children = [
    control.cloudtrail_trail_validation_enabled,
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_patch_compliant
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_si_12" {
  title       = "Information Handling and Retention (SI-12)"
  description = "The organization handles and retains information within the information system and information output from the system in accordance with applicable federal laws, Executive Orders, directives, policies, regulations, standards, and operational requirements."
  children = [
    control.cloudwatch_log_group_retention_period_365,
    control.dynamodb_table_in_backup_plan,
    control.dynamodb_table_point_in_time_recovery_enabled,
    control.ebs_volume_in_backup_plan,
    control.efs_file_system_in_backup_plan,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.rds_db_instance_backup_enabled,
    control.rds_db_instance_in_backup_plan,
    control.s3_bucket_versioning_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
