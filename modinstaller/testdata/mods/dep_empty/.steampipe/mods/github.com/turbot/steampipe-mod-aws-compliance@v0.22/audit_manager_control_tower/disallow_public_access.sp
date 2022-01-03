locals {
  audit_manager_control_tower_disallow_public_access_common_tags = merge(local.audit_manager_control_tower_common_tags, {
    control_set = "disallow_public_access"
  })
}

benchmark "audit_manager_control_tower_disallow_public_access" {
  title         = "Disallow Public Access"
  description   = "This benchmark checks if RDS instances, snapshots and S3 buckets are not publicly accessible."
  children = [
    benchmark.audit_manager_control_tower_disallow_public_access_4_0_1,
    benchmark.audit_manager_control_tower_disallow_public_access_4_0_2,
    benchmark.audit_manager_control_tower_disallow_public_access_4_1_1,
    benchmark.audit_manager_control_tower_disallow_public_access_4_1_2
  ]
  tags          = local.audit_manager_control_tower_disallow_public_access_common_tags
}

benchmark "audit_manager_control_tower_disallow_public_access_4_0_1" {
  title         = "4.0.1 - Disallow public access to RDS database instances"
  description   = "Disallow public access to RDS database instances - Checks whether the Amazon Relational Database Service (RDS) instances are not publicly accessible. The rule is non-compliant if the publiclyAccessible field is true in the instance configuration item."
  children = [
    control.rds_db_instance_prohibit_public_access
  ]

  tags = merge(local.audit_manager_control_tower_disallow_public_access_common_tags, {
    audit_manager_control_tower_item_id = "4.0.1"
  })
}

benchmark "audit_manager_control_tower_disallow_public_access_4_0_2" {
  title         = "4.0.2 - Disallow public access to RDS database snapshots"
  description   = "Disallow public access to RDS database snapshots - Checks if Amazon Relational Database Service (Amazon RDS) snapshots are public. The rule is non-compliant if any existing and new Amazon RDS snapshots are public."
  children = [
    control.rds_db_snapshot_prohibit_public_access
  ]

  tags = merge(local.audit_manager_control_tower_disallow_public_access_common_tags, {
    audit_manager_control_tower_item_id = "4.0.2"
  })
}

benchmark "audit_manager_control_tower_disallow_public_access_4_1_1" {
  title         = "4.1.1 - Disallow public read access to S3 buckets"
  description   = "Disallow public read access to S3 buckets - Checks that your S3 buckets do not allow public read access."
  children = [
    control.s3_bucket_restrict_public_read_access
  ]

  tags = merge(local.audit_manager_control_tower_disallow_public_access_common_tags, {
    audit_manager_control_tower_item_id = "4.1.1"
  })
}

benchmark "audit_manager_control_tower_disallow_public_access_4_1_2" {
  title         = "4.1.2 - Disallow public write access to S3 buckets"
  description   = "Disallow public write access to S3 buckets - Checks that your S3 buckets do not allow public write access."
  children = [
    control.s3_bucket_restrict_public_write_access
  ]

  tags = merge(local.audit_manager_control_tower_disallow_public_access_common_tags, {
    audit_manager_control_tower_item_id = "4.1.2"
  })
}