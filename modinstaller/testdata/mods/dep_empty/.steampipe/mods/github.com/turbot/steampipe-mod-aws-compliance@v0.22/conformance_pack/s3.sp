locals {
  conformance_pack_s3_common_tags = {
    service = "s3"
  }
}

control "s3_bucket_cross_region_replication_enabled" {
  title       = "S3 bucket cross-region replication should enabled"
  description = "Amazon Simple Storage Service (Amazon S3) Cross-Region Replication (CRR) supports maintaining adequate capacity and availability."
  sql         = query.s3_bucket_cross_region_replication_enabled.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "s3_bucket_default_encryption_enabled" {
  title       = "S3 bucket default encryption should be enabled"
  description = "To help protect data at rest, ensure encryption is enabled for your Amazon Simple Storage Service (Amazon S3) buckets."
  sql         = query.s3_bucket_default_encryption_enabled.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "s3_bucket_enforces_ssl" {
  title       = "S3 buckets should enforce SSL"
  description = "To help protect data in transit, ensure that your Amazon Simple Storage Service (Amazon S3) buckets require requests to use Secure Socket Layer (SSL)."
  sql         = query.s3_bucket_enforces_ssl.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "s3_bucket_logging_enabled" {
  title       = "S3 bucket logging should be enabled"
  description = "Amazon Simple Storage Service (Amazon S3) server access logging provides a method to monitor the network for potential cybersecurity events."
  sql         = query.s3_bucket_logging_enabled.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "s3_bucket_object_lock_enabled" {
  title       = "S3 bucket object lock should be enabled"
  description = "Ensure that your Amazon Simple Storage Service (Amazon S3) bucket has lock enabled, by default."
  sql         = query.s3_bucket_object_lock_enabled.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}

control "s3_bucket_restrict_public_read_access" {
  title       = "S3 buckets should prohibit public read access"
  description = "Manage access to resources in the AWS Cloud by only allowing authorized users, processes, and devices access to Amazon Simple Storage Service (Amazon S3) buckets."
  sql         = query.s3_bucket_restrict_public_read_access.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    audit_manager_control_tower = "true"
    hipaa                       = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
    rbi_cyber_security          = "true"
    soc_2                       = "true"
  })
}

control "s3_bucket_restrict_public_write_access" {
  title       = "S3 buckets should prohibit public write access"
  description = "Manage access to resources in the AWS Cloud by only allowing authorized users, processes, and devices access to Amazon Simple Storage Service (Amazon S3) buckets."
  sql         = query.s3_bucket_restrict_public_write_access.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    audit_manager_control_tower = "true"
    hipaa                       = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
    rbi_cyber_security          = "true"
  })
}

control "s3_bucket_versioning_enabled" {
  title       = "S3 bucket versioning should be enabled"
  description = "Amazon Simple Storage Service (Amazon S3) bucket versioning helps keep multiple variants of an object in the same Amazon S3 bucket."
  sql         = query.s3_bucket_versioning_enabled.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    audit_manager_control_tower = "true"
    hipaa                       = "true"
    nist_csf                    = "true"
    rbi_cyber_security          = "true"
    soc_2                       = "true"
  })
}

control "s3_public_access_block_account" {
  title       = "S3 public access should be blocked at account level"
  description = "Manage access to resources in the AWS Cloud by ensuring that Amazon Simple Storage Service (Amazon S3) buckets cannot be publicly accessed."
  sql         = query.s3_public_access_block_account.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
  })
}

control "s3_public_access_block_bucket_account" {
  title       = "S3 public access should be blocked at account and bucket levels"
  description = "Manage access to resources in the AWS Cloud by ensuring that Amazon Simple Storage Service (Amazon S3) buckets cannot be publicly accessed."
  sql         = query.s3_public_access_block_bucket_account.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "s3_bucket_default_encryption_enabled_kms" {
  title       = "S3 bucket default encryption should be enabled with KMS"
  description = "To help protect data at rest, ensure encryption is enabled for your Amazon Simple Storage Service (Amazon S3) buckets."
  sql         = query.s3_bucket_default_encryption_enabled_kms.sql

  tags = merge(local.conformance_pack_s3_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    rbi_cyber_security = "true"
  })
}