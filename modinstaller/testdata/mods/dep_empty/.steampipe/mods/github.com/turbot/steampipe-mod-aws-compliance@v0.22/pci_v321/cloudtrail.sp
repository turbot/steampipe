locals {
  pci_v321_cloudtrail_common_tags = merge(local.pci_v321_common_tags, {
    service = "cloudtrail"
  })
}

benchmark "pci_v321_cloudtrail" {
  title         = "CloudTrail"
  documentation = file("./pci_v321/docs/pci_v321_cloudtrail.md")
  children = [
    control.pci_v321_cloudtrail_1,
    control.pci_v321_cloudtrail_2,
    control.pci_v321_cloudtrail_3,
    control.pci_v321_cloudtrail_4
  ]
  tags          = local.pci_v321_cloudtrail_common_tags
}

control "pci_v321_cloudtrail_1" {
  title         = "1 CloudTrail logs should be encrypted at rest using AWS KMS CMKs"
  description   = "This control checks whether AWS CloudTrail is configured to use the server-side encryption (SSE) AWS KMS customer master key (CMK) encryption. If you are only using the default encryption option, you can choose to disable this check."
  severity      = "medium"
  sql           = query.cloudtrail_trail_logs_encrypted_with_kms_cmk.sql
  documentation = file("./pci_v321/docs/pci_v321_cloudtrail_1.md")

  tags = merge(local.pci_v321_cloudtrail_common_tags, {
    pci_item_id      = "cloudtrail_1"
    pci_requirements = "3.4"
  })
}

control "pci_v321_cloudtrail_2" {
  title         = "2 CloudTrail should be enabled"
  description   = "This control checks whether CloudTrail is enabled in your AWS account. However, some AWS services do not enable logging of all APIs and events. You should implement any additional audit trails other than CloudTrail and review the documentation for each service."
  severity      = "high"
  sql           = query.cloudtrail_enabled_all_regions.sql
  documentation = file("./pci_v321/docs/pci_v321_cloudtrail_2.md")

  tags = merge(local.pci_v321_cloudtrail_common_tags, {
    pci_item_id      = "cloudtrail_2"
    pci_requirements = "10.1,10.2.1,10.2.2,10.2.3,10.2.4,10.2.5,10.2.6,10.2.7,10.3.1,10.3.2,10.3.3,10.3.4,10.3.5,10.3.6"
  })
}

control "pci_v321_cloudtrail_3" {
  title         = "3 CloudTrail log file validation should be enabled"
  description   = "This control checks whether CloudTrail log file validation is enabled."
  severity      = "low"
  sql           = query.cloudtrail_trail_validation_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_cloudtrail_3.md")

  tags = merge(local.pci_v321_cloudtrail_common_tags, {
    pci_item_id      = "cloudtrail_3"
    pci_requirements = "10.5.2,10.5.5"
  })
}

control "pci_v321_cloudtrail_4" {
  title         = "4 CloudTrail trails should be integrated with CloudWatch Logs"
  description   = "This control checks whether CloudTrail trails are configured to send logs to CloudWatch Logs."
  severity      = "low"
  sql           = query.cloudtrail_trail_integrated_with_logs.sql
  documentation = file("./pci_v321/docs/pci_v321_cloudtrail_4.md")

  tags = merge(local.pci_v321_cloudtrail_common_tags, {
    pci_item_id      = "cloudtrail_4"
    pci_requirements = "10.5.3"
  })
}