benchmark "nist_csf_rs" {
  title       = "Respond (RS)"
  description = "Develop and implement appropriate activities to take action regarding a detected cybersecurity incident."

  children = [
    benchmark.nist_csf_rs_an,
    benchmark.nist_csf_rs_mi,
    benchmark.nist_csf_rs_rp
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rs_an" {
  title       = "Analysis (RS.AN)"
  description = "Analysis is conducted to ensure effective response and support recovery activities."

  children = [
    benchmark.nist_csf_rs_an_2
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rs_an_2" {
  title       = "RS.AN-2"
  description = "The impact of the incident is understood."

  children = [
    control.guardduty_finding_archived
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rs_mi" {
  title       = "Mitigation (RS.MI)"
  description = "Activities are performed to prevent expansion of an event, mitigate its effects, and eradicate the incident."

  children = [
    benchmark.nist_csf_rs_mi_3
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rs_mi_3" {
  title       = "RS.MI-3"
  description = "Newly identified vulnerabilities are mitigated or documented as accepted risks."

  children = [
    control.guardduty_finding_archived
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rs_rp" {
  title       = "Response Planning (RS.RP)"
  description = "Response processes and procedures are run and maintained, to ensure timely response to detected cybersecurity events."

  children = [
    benchmark.nist_csf_rs_rp_1
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rs_rp_1" {
  title       = "RS.RP-1"
  description = "Response plan is executed during or after an incident."

  children = [
    control.backup_plan_min_retention_35_days,
    control.backup_recovery_point_encryption_enabled,
    control.backup_recovery_point_manual_deletion_disabled,
    control.dynamodb_table_auto_scaling_enabled,
    control.dynamodb_table_in_backup_plan,
    control.dynamodb_table_point_in_time_recovery_enabled,
    control.dynamodb_table_protected_by_backup_plan,
    control.ebs_volume_in_backup_plan,
    control.ebs_volume_protected_by_backup_plan,
    control.ec2_instance_protected_by_backup_plan,
    control.efs_file_system_in_backup_plan,
    control.efs_file_system_protected_by_backup_plan,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.elb_application_lb_deletion_protection_enabled,
    control.elb_classic_lb_cross_zone_load_balancing_enabled,
    control.fsx_file_system_protected_by_backup_plan,
    control.rds_db_cluster_aurora_protected_by_backup_plan,
    control.rds_db_instance_backup_enabled,
    control.rds_db_instance_in_backup_plan,
    control.rds_db_instance_multiple_az_enabled,
    control.rds_db_instance_protected_by_backup_plan,
    control.redshift_cluster_automatic_snapshots_min_7_days,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled,
    control.vpc_vpn_tunnel_up
  ]

  tags = local.nist_csf_common_tags
}