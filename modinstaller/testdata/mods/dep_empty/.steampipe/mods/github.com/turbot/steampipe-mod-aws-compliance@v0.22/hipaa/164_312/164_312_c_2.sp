benchmark "hipaa_164_312_c_2" {
  title       = "164.312(c)(2) Mechanism to authenticate electronic protected health information"
  description = "Implement electronic mechanisms to corroborate that electronic protected health information has not been altered or destroyed in an unauthorized manner."
  children = [
    control.cloudtrail_trail_logs_encrypted_with_kms_cmk,
    control.cloudtrail_trail_validation_enabled,
    control.ebs_attached_volume_encryption_enabled,
    control.s3_bucket_default_encryption_enabled,
    control.s3_bucket_enforces_ssl,
    control.s3_bucket_object_lock_enabled,
    control.s3_bucket_versioning_enabled,
    control.vpc_flow_logs_enabled
  ]

  tags = merge(local.hipaa_164_312_common_tags, {
    hipaa_item_id = "164_312_c_2"
  })
}