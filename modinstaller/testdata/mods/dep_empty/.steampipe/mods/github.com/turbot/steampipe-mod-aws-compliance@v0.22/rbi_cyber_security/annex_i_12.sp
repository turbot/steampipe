benchmark "rbi_cyber_security_annex_i_12" {
  title       = "Annex I (12)"
  description = "Take periodic back up of the important data and store this data ‘off line’ (i.e., transferring important files to a storage device that can be detached from a computer/system after copying all the files)."

  children = [
    control.dynamodb_table_in_backup_plan,
    control.dynamodb_table_point_in_time_recovery_enabled,
    control.ebs_volume_in_backup_plan,
    control.efs_file_system_in_backup_plan,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.rds_db_instance_backup_enabled,
    control.rds_db_instance_in_backup_plan,
    control.redshift_cluster_automatic_snapshots_min_7_days,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_12"
  })
}
