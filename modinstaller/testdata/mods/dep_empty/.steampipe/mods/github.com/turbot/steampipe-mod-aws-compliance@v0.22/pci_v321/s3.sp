locals {
  pci_v321_s3_common_tags = merge(local.pci_v321_common_tags, {
    service = "s3"
  })
}

benchmark "pci_v321_s3" {
  title         = "S3"
  documentation = file("./pci_v321/docs/pci_v321_s3.md")
  children = [
    control.pci_v321_s3_1,
    control.pci_v321_s3_2,
    control.pci_v321_s3_3,
    control.pci_v321_s3_4,
    control.pci_v321_s3_5,
    control.pci_v321_s3_6,
  ]
  tags = local.pci_v321_s3_common_tags
}

control "pci_v321_s3_1" {
  title         = "1 S3 buckets should prohibit public write access"
  description   = "This control checks whether your S3 buckets allow public write access by evaluating the Block Public Access settings, the bucket policy, and the bucket access control list (ACL). It does not check for write access to the bucket by internal principals, such as IAM roles. You should ensure that access to the bucket is restricted to authorized principals only."
  severity      = "critical"
  sql           = query.s3_bucket_restrict_public_write_access.sql
  documentation = file("./pci_v321/docs/pci_v321_s3_1.md")

  tags = merge(local.pci_v321_s3_common_tags, {
    pci_item_id      = "s3_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6,7.2.1"
  })
}

control "pci_v321_s3_2" {
  title         = "2 S3 buckets should prohibit public read access"
  description   = "This control checks whether your S3 buckets allow public read access by evaluating the Block Public Access settings, the bucket policy, and the bucket access control list (ACL). Unless you explicitly require everyone on the internet to be able to write to your S3 bucket, you should ensure that your S3 bucket is not publicly writable. It does not check for read access to the bucket by internal principals, such as IAM roles. You should ensure that access to the bucket is restricted to authorized principals only."
  severity      = "critical"
  sql           = query.s3_bucket_restrict_public_read_access.sql
  documentation = file("./pci_v321/docs/pci_v321_s3_2.md")

  tags = merge(local.pci_v321_s3_common_tags, {
    pci_item_id      = "s3_2"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.6,7.2.1"
  })
}

control "pci_v321_s3_3" {
  title         = "3 S3 buckets should have cross-region replication enabled"
  description   = "This control checks whether S3 buckets have cross-region replication enabled. PCI DSS does not require data replication or highly available configurations. However, this check aligns with AWS best practices for this control."
  severity      = "low"
  sql           = query.s3_bucket_cross_region_replication_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_s3_3.md")

  tags = merge(local.pci_v321_s3_common_tags, {
    pci_item_id      = "s3_3"
    pci_requirements = "2.2"
  })
}

control "pci_v321_s3_4" {
  title         = "4 S3 buckets should have server-side encryption enabled"
  description   = "This control checks that your Amazon S3 bucket either has Amazon S3 default encryption enabled or that the S3 bucket policy explicitly denies put-object requests without server-side encryption. When you set default encryption on a bucket, all new objects stored in the bucket are encrypted when they are stored, including clear text PAN data. Server-side encryption for all of the objects stored in a bucket can also be enforced using a bucket policy."
  severity      = "medium"
  sql           = query.s3_bucket_default_encryption_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_s3_4.md")

  tags = merge(local.pci_v321_s3_common_tags, {
    pci_item_id      = "s3_4"
    pci_requirements = "3.4"
  })
}

control "pci_v321_s3_5" {
  title         = "5 S3 buckets should require requests to use Secure Socket Layer"
  description   = "This control checks whether Amazon S3 buckets have policies that require requests to use Secure Socket Layer (SSL). S3 buckets should have policies that require all requests (Action: S3:*)to only accept transmission of data over HTTPS in the S3 resource policy, indicated by the condition key aws:SecureTransport."
  severity      = "medium"
  sql           = query.s3_bucket_enforces_ssl.sql
  documentation = file("./pci_v321/docs/pci_v321_s3_4.md")

  tags = merge(local.pci_v321_s3_common_tags, {
    pci_item_id      = "s3_5"
    pci_requirements = "4.1"

  })
}

control "pci_v321_s3_6" {
  title         = "6 S3 Block Public Access setting should be enabled"
  description   = "This control checks whether the following public access block settings are configured at the account level. The control passes if all of the public access block settings are set to true. The control fails if any of the settings are set to false, or if any of the settings are not configured. When the settings do not have a value, the AWS Config rule cannot complete its evaluation."
  severity      = "medium"
  sql           = query.s3_public_access_block_bucket_account.sql
  documentation = file("./pci_v321/docs/pci_v321_s3_6.md")

  tags = merge(local.pci_v321_s3_common_tags, {
    pci_item_id      = "s3_6"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6"
  })
}