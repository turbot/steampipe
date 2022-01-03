locals {
  foundational_security_cloudtrail_common_tags = merge(local.foundational_security_common_tags, {
    service = "cloudtrail"
  })
}

benchmark "foundational_security_cloudtrail" {
  title         = "CloudTrail"
  documentation = file("./foundational_security/docs/foundational_security_cloudtrail.md")
  children = [
    control.foundational_security_cloudtrail_1,
    control.foundational_security_cloudtrail_2,
    control.foundational_security_cloudtrail_4,
    control.foundational_security_cloudtrail_5
  ]
  tags          = local.foundational_security_cloudtrail_common_tags
}

control "foundational_security_cloudtrail_1" {
  title         = "1 CloudTrail should be enabled and configured with at least one multi-Region trail"
  description   = "This control checks that there is at least one multi-Region CloudTrail trail."
  severity      = "high"
  sql           = query.cloudtrail_multi_region_trail_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudtrail_2.md")

  tags = merge(local.foundational_security_cloudtrail_common_tags, {
    foundational_security_item_id  = "cloudtrail_1"
    foundational_security_category = "logging"
  })
}

control "foundational_security_cloudtrail_2" {
  title         = "2 CloudTrail should have encryption at rest enabled"
  description   = "This control checks whether CloudTrail is configured to use the server-side encryption (SSE) AWS Key Management Service customer master key (CMK) encryption. The check passes if the KmsKeyId is defined."
  severity      = "medium"
  sql           = query.cloudtrail_trail_logs_encrypted_with_kms_cmk.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudtrail_2.md")

  tags = merge(local.foundational_security_cloudtrail_common_tags, {
    foundational_security_item_id  = "cloudtrail_2"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}

control "foundational_security_cloudtrail_4" {
  title         = "4 Ensure CloudTrail log file validation is enabled"
  description   = "This control checks whether log file integrity validation is enabled on a CloudTrail trail. CloudTrail log file validation creates a digitally signed digest file that contains a hash of each log that CloudTrail writes to Amazon S3. You can use these digest files to determine whether a log file was changed, deleted, or unchanged after CloudTrail delivered the log."
  severity      = "low"
  sql           = query.cloudtrail_trail_validation_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudtrail_4.md")

  tags = merge(local.foundational_security_cloudtrail_common_tags, {
    foundational_security_item_id  = "cloudtrail_4"
    foundational_security_category = "data_integrity"
  })
}

control "foundational_security_cloudtrail_5" {
  title         = "5 Ensure CloudTrail trails are integrated with Amazon CloudWatch Logs"
  description   = "This control checks whether CloudTrail trails are configured to send logs to CloudWatch Logs. The control fails if the CloudWatchLogsLogGroupArn property of the trail is empty."
  severity      = "low"
  sql           = query.cloudtrail_trail_integrated_with_logs.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudtrail_5.md")

  tags = merge(local.foundational_security_cloudtrail_common_tags, {
    foundational_security_item_id  = "cloudtrail_5"
    foundational_security_category = "logging"
  })
}