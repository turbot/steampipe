locals {
  cis_v140_2_common_tags = merge(local.cis_v140_common_tags, {
    cis_section_id = "2"
  })
}

locals {
  cis_v140_2_1_common_tags = merge(local.cis_v140_2_common_tags, {
    cis_section_id = "2.1"
  })
  cis_v140_2_2_common_tags = merge(local.cis_v140_2_common_tags, {
    cis_section_id = "2.2"
  })
  cis_v140_2_3_common_tags = merge(local.cis_v140_2_common_tags, {
    cis_section_id = "2.3"
  })
}

benchmark "cis_v140_2" {
  title         = "2 Storage"
  documentation = file("./cis_v140/docs/cis_v140_2.md")
  children = [
    benchmark.cis_v140_2_1,
    benchmark.cis_v140_2_2,
    benchmark.cis_v140_2_3
  ]
  tags = local.cis_v140_2_common_tags
}

benchmark "cis_v140_2_1" {
  title         = "2.1 Simple Storage Service (S3)"
  documentation = file("./cis_v140/docs/cis_v140_2_1.md")
  children = [
    control.cis_v140_2_1_1,
    control.cis_v140_2_1_2,
    control.cis_v140_2_1_3,
    control.cis_v140_2_1_4,
    control.cis_v140_2_1_5
  ]
  tags = local.cis_v140_2_1_common_tags
}

control "cis_v140_2_1_1" {
  title         = "2.1.1 Ensure all S3 buckets employ encryption-at-rest"
  description   = "Amazon S3 provides a variety of no, or low, cost encryption options to protect data at rest."
  documentation = file("./cis_v140/docs/cis_v140_2_1_1.md")
  sql           = query.s3_bucket_default_encryption_enabled.sql

  tags = merge(local.cis_v140_2_1_common_tags, {
    cis_item_id = "2.1.1"
    cis_level   = "2"
    cis_type    = "manual"
    service     = "s3"
  })
}

control "cis_v140_2_1_2" {
  title         = "2.1.2 Ensure S3 Bucket Policy is set to deny HTTP requests"
  description   = "At the Amazon S3 bucket level, you can configure permissions through a bucket policy making the objects accessible only through HTTPS."
  documentation = file("./cis_v140/docs/cis_v140_2_1_2.md")
  sql           = query.s3_bucket_enforces_ssl.sql

  tags = merge(local.cis_v140_2_1_common_tags, {
    cis_item_id = "2.1.2"
    cis_level   = "2"
    cis_type    = "manual"
    service     = "s3"
  })
}

control "cis_v140_2_1_3" {
  title         = "2.1.3 Ensure MFA Delete is enabled on S3 buckets"
  description   = "Once MFA Delete is enabled on your sensitive and classified S3 bucket it requires the user to have two forms of authentication."
  documentation = file("./cis_v140/docs/cis_v140_2_1_3.md")
  sql           = query.s3_bucket_mfa_delete_enabled.sql

  tags = merge(local.cis_v140_2_1_common_tags, {
    cis_item_id = "2.1.3"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "s3"
  })
}

control "cis_v140_2_1_4" {
  title         = "2.1.4 Ensure all data in Amazon S3 has been discovered, classified and secured when required"
  description   = "Amazon S3 buckets can contain sensitive data, that for security purposes should be discovered, monitored, classified and protected. Macie along with other 3rd party tools can automatically provide an inventory of Amazon S3 buckets."
  documentation = file("./cis_v140/docs/cis_v140_2_1_4.md")
  sql           = query.manual_control.sql

  tags = merge(local.cis_v140_2_1_common_tags, {
    cis_item_id = "2.1.4"
    cis_level   = "2"
    cis_type    = "manual"
    service     = "s3"
  })
}

control "cis_v140_2_1_5" {
  title         = "2.1.5 Ensure that S3 Buckets are configured with 'Block public access (bucket settings)'"
  description   = "Amazon S3 provides Block public access (bucket settings) and Block public access (account settings) to help you manage public access to Amazon S3 resources. By default, S3 buckets and objects are created with public access disabled. However, an IAM principle with sufficient S3 permissions can enable public access at the bucket and/or object level. While enabled, Block public access (bucket settings) prevents an individual bucket, and its contained objects, from becoming publicly accessible. Similarly, Block public access (account settings) prevents all buckets, and contained objects, from becoming publicly accessible across the entire account."
  documentation = file("./cis_v140/docs/cis_v140_2_1_5.md")
  sql           = query.s3_public_access_block_bucket_account.sql

  tags = merge(local.cis_v140_2_1_common_tags, {
    cis_item_id = "2.1.5"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "s3"
  })
}

benchmark "cis_v140_2_2" {
  title         = "2.2 Elastic Compute Cloud (EC2)"
  documentation = file("./cis_v140/docs/cis_v140_2_2.md")
  children = [
    control.cis_v140_2_2_1
  ]
  tags = local.cis_v140_2_2_common_tags
}

control "cis_v140_2_2_1" {
  title         = "2.2.1 Ensure EBS volume encryption is enabled"
  description   = "Elastic Compute Cloud (EC2) supports encryption at rest when using the Elastic Block Store (EBS) service. While disabled by default, forcing encryption at EBS volume creation is supported."
  documentation = file("./cis_v140/docs/cis_v140_2_2_1.md")
  sql           = query.ebs_volume_encryption_at_rest_enabled.sql

  tags = merge(local.cis_v140_2_2_common_tags, {
    cis_item_id = "2.2.1"
    cis_level   = "1"
    cis_type    = "manual"
    service     = "ebs"
  })
}

benchmark "cis_v140_2_3" {
  title         = "2.3 Relational Database Service (RDS)"
  documentation = file("./cis_v140/docs/cis_v140_2_3.md")
  children = [
    control.cis_v140_2_3_1
  ]
  tags = local.cis_v140_2_3_common_tags
}

control "cis_v140_2_3_1" {
  title         = "2.3.1 Ensure that encryption is enabled for RDS Instances"
  description   = "Amazon RDS encrypted DB instances use the industry standard AES-256 encryption algorithm to encrypt your data on the server that hosts your Amazon RDS DB instances. After your data is encrypted, Amazon RDS handles authentication of access and decryption of your data transparently with a minimal impact on performance."
  documentation = file("./cis_v140/docs/cis_v140_2_3_1.md")
  sql           = query.rds_db_instance_encryption_at_rest_enabled.sql

  tags = merge(local.cis_v140_2_3_common_tags, {
    cis_item_id = "2.3.1"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "rds"
  })
}
