locals {
  foundational_security_s3_common_tags = merge(local.foundational_security_common_tags, {
    service = "s3"
  })
}

benchmark "foundational_security_s3" {
  title         = "S3"
  documentation = file("./foundational_security/docs/foundational_security_s3.md")
  children = [
    control.foundational_security_s3_1,
    control.foundational_security_s3_2,
    control.foundational_security_s3_3,
    control.foundational_security_s3_4,
    control.foundational_security_s3_5,
    control.foundational_security_s3_6,
    control.foundational_security_s3_8
  ]
  tags          = local.foundational_security_s3_common_tags
}

control "foundational_security_s3_1" {
  title         = "1 S3 Block Public Access setting should be enabled"
  description   = "This control checks whether the following Amazon S3 public access block settings are configured at the account level"
  severity      = "medium"
  sql           = query.s3_public_access_block_account.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_1.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_1"
    foundational_security_category = "secure_network_configuration"
  })
}

control "foundational_security_s3_2" {
  title         = "2 S3 buckets should prohibit public read access"
  description   = "This control checks whether your S3 buckets allow public read access. It evaluates the Block Public Access settings, the bucket policy, and the bucket access control list (ACL)."
  severity      = "critical"
  sql           = query.s3_bucket_restrict_public_read_access.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_2.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_2"
    foundational_security_category = "secure_network_configuration"
  })
}

control "foundational_security_s3_3" {
  title         = "3 S3 buckets should prohibit public write access"
  description   = "This control checks whether your S3 buckets allow public write access. It evaluates the block public access settings, the bucket policy, and the bucket access control list (ACL)."
  severity      = "critical"
  sql           = query.s3_bucket_restrict_public_write_access.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_3.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_3"
    foundational_security_category = "secure_network_configuration"
  })
}

control "foundational_security_s3_4" {
  title         = "4 S3 buckets should have server-side encryption enabled"
  description   = "This control checks that your S3 bucket either has Amazon S3 default encryption enabled or that the S3 bucket policy explicitly denies put-object requests without server-side encryption."
  severity      = "medium"
  sql           = query.s3_bucket_default_encryption_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_4.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_4"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}

control "foundational_security_s3_5" {
  title         = "5 S3 buckets should require requests to use Secure Socket Layer"
  description   = "This control checks whether S3 buckets have policies that require requests to use Secure Socket Layer (SSL). S3 buckets should have policies that require all requests (Action: S3:*)to only accept transmission of data over HTTPS in the S3 resource policy, indicated by the condition key aws:SecureTransport."
  severity      = "medium"
  sql           = query.s3_bucket_enforces_ssl.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_5.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_5"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_s3_6" {
  title         = "6 Amazon S3 permissions granted to other AWS accounts in bucket policies should be restricted"
  description   = "This control checks whether the S3 bucket policy prevents principals from other AWS accounts from performing denied actions on resources in the S3 bucket."
  severity      = "high"
  sql           = query.s3_bucket_policy_restricts_cross_account_permission_changes.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_6.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_6"
    foundational_security_category = "sensitive_api_operations_actions_restricted"
  })
}

control "foundational_security_s3_8" {
  title         = "8 S3 Block Public Access setting should be enabled at the bucket level"
  description   = "This control checks whether S3 buckets have bucket-level public access blocks applied."
  severity      = "high"
  sql           = query.s3_bucket_public_access_blocked.sql
  documentation = file("./foundational_security/docs/foundational_security_s3_8.md")

  tags = merge(local.foundational_security_s3_common_tags, {
    foundational_security_item_id  = "s3_8"
    foundational_security_category = "access_control"
  })
}