benchmark "nist_csf_rc" {
  title       = "Recover (RC)"
  description = "Develop and implement appropriate activities to maintain plans for resilience and to restore any capabilities or services that were impaired due to a cybersecurity incident."

  children = [
    benchmark.nist_csf_rc_rp
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rc_rp" {
  title       = "Recovery Planning (RC.RP)"
  description = "Recovery processes and procedures are executed and maintained to ensure timely restoration of systems or assets affected by cybersecurity events."

  children = [
    benchmark.nist_csf_rc_rp_1
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_rc_rp_1" {
  title       = "RC.RP-1"
  description = "Recovery plan is executed during or after a cybersecurity incident."

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