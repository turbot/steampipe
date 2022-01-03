locals {
  soc_2_cc_7_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "cc7"
  })
}

benchmark "soc_2_cc_7" {
  title       = "CC7.0 - System Operations"
  description = "The criteria relevant to how an entity (i) manages the operation of system(s) and (ii) detects and mitigates processing deviations including logical and physical security deviations."

  children = [
    benchmark.soc_2_cc_7_1,
    benchmark.soc_2_cc_7_2,
    benchmark.soc_2_cc_7_3,
    benchmark.soc_2_cc_7_4,
    benchmark.soc_2_cc_7_5
  ]

  tags = local.soc_2_cc_7_common_tags
}

benchmark "soc_2_cc_7_1" {
  title         = "CC7.1 To meet its objectives, the entity uses detection and monitoring procedures to identify (1) changes to configurations that result in the introduction of new vulnerabilities, and (2) susceptibilities to newly discovered vulnerabilities"
  documentation = file("./soc2/docs/cc_7_1.md")

  children = [
    control.guardduty_enabled,
    control.securityhub_enabled,
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant
  ]

  tags = merge(local.soc_2_cc_7_common_tags, {
    soc_2_item_id = "7.1"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_7_2" {
  title         = "CC7.2 The entity monitors system components and the operation of those components for anomalies that are indicative of malicious acts, natural disasters, and errors affecting the entity's ability to meet its objectives; anomalies are analyzed to determine whether they represent security events"
  documentation = file("./soc2/docs/cc_7_2.md")

  children = [
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.lambda_function_dead_letter_queue_configured,
    control.elb_application_classic_lb_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.rds_db_instance_logging_enabled,
    control.wafv2_web_acl_logging_enabled,
    control.cloudtrail_trail_enabled,
    control.codebuild_project_plaintext_env_variables_no_sensitive_aws_values,
    control.securityhub_enabled,
    control.cloudwatch_log_group_retention_period_365,
    control.cloudtrail_multi_region_trail_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.vpc_flow_logs_enabled,
    control.ec2_instance_detailed_monitoring_enabled,
    control.codebuild_project_source_repo_oauth_configured,
    control.guardduty_enabled,
    control.apigateway_stage_logging_enabled,
    control.lambda_function_concurrent_execution_limit_configured,
    control.vpc_security_group_restrict_ingress_ssh_all
  ]

  tags = merge(local.soc_2_cc_7_common_tags, {
    soc_2_item_id = "7.2"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_7_3" {
  title         = "CC7.3 The entity evaluates security events to determine whether they could or have resulted in a failure of the entity to meet its objectives (security incidents) and, if so, takes actions to prevent or address such failures"
  documentation = file("./soc2/docs/cc_7_3.md")

  children = [
    control.log_group_encryption_at_rest_enabled,
    control.cloudtrail_trail_validation_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.guardduty_enabled,
    control.apigateway_stage_logging_enabled,
    control.lambda_function_dead_letter_queue_configured,
    control.rds_db_instance_logging_enabled,
    control.securityhub_enabled,
    control.cloudwatch_alarm_action_enabled,
    control.elb_application_classic_lb_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.cloudwatch_log_group_retention_period_365,
    control.vpc_flow_logs_enabled,
    control.guardduty_finding_archived,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = merge(local.soc_2_cc_7_common_tags, {
    soc_2_item_id = "7.3"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_7_4" {
  title         = "CC7.4 The entity responds to identified security incidents by executing a defined incident response program to understand, contain, remediate, and communicate security incidents, as appropriate"
  documentation = file("./soc2/docs/cc_7_4.md")

  children = [
    control.backup_plan_min_retention_35_days,
    control.backup_recovery_point_encryption_enabled,
    control.backup_recovery_point_manual_deletion_disabled,
    control.cloudwatch_alarm_action_enabled,
    control.dynamodb_table_in_backup_plan,
    control.dynamodb_table_point_in_time_recovery_enabled,
    control.dynamodb_table_protected_by_backup_plan,
    control.ebs_volume_in_backup_plan,
    control.ebs_volume_protected_by_backup_plan,
    control.ec2_instance_ebs_optimized,
    control.ec2_instance_protected_by_backup_plan,
    control.efs_file_system_in_backup_plan,
    control.efs_file_system_protected_by_backup_plan,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.fsx_file_system_protected_by_backup_plan,
    control.guardduty_enabled,
    control.guardduty_finding_archived,
    control.lambda_function_dead_letter_queue_configured,
    control.rds_db_cluster_aurora_protected_by_backup_plan,
    control.rds_db_instance_backup_enabled,
    control.rds_db_instance_in_backup_plan,
    control.rds_db_instance_protected_by_backup_plan,
    control.redshift_cluster_automatic_snapshots_min_7_days,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled,
    control.securityhub_enabled
  ]

  tags = merge(local.soc_2_cc_7_common_tags, {
    soc_2_item_id = "7.4"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_7_5" {
  title         = "CC7.5 The entity identifies, develops, and implements activities to recover from identified security incidents"
  documentation = file("./soc2/docs/cc_7_5.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_7_common_tags, {
    soc_2_item_id = "7.5"
    soc_2_type    = "manual"
  })
}