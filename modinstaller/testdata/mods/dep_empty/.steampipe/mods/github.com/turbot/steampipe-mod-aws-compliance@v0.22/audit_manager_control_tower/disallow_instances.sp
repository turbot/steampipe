locals {
  audit_manager_control_tower_disallow_instances_common_tags = merge(local.audit_manager_control_tower_common_tags, {
    control_set = "disallow_instances"
  })
}

benchmark "audit_manager_control_tower_disallow_instances" {
  title        = "Disallow Instances"
  description  = "This benchmark checks if RDS storage is encrypted and S3 bucket's versioning is enabled."
  children = [
    benchmark.audit_manager_control_tower_disallow_instances_5_0_1,
    benchmark.audit_manager_control_tower_disallow_instances_5_1_1
  ]
  tags          = local.audit_manager_control_tower_disallow_instances_common_tags
}

benchmark "audit_manager_control_tower_disallow_instances_5_0_1" {
  title         = "5.0.1 - Disallow RDS database instances that are not storage encrypted"
  description   = "Disallow RDS database instances that are not storage encrypted - Checks whether storage encryption is enabled for your RDS DB instances."
  children = [
    control.rds_db_snapshot_encrypted_at_rest
  ]

  tags = merge(local.audit_manager_control_tower_disallow_instances_common_tags, {
    audit_manager_control_tower_item_id  = "5.0.1"
  })
}

benchmark "audit_manager_control_tower_disallow_instances_5_1_1" {
  title         = "5.1.1 - Disallow S3 buckets that are not versioning enabled"
  description   = "Disallow S3 buckets that are not versioning enabled - Checks whether versioning is enabled for your S3 buckets."
  children = [
    control.s3_bucket_versioning_enabled
  ]

  tags = merge(local.audit_manager_control_tower_disallow_instances_common_tags, {
    audit_manager_control_tower_item_id  = "5.1.1"
  })
}