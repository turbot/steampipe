locals {
  conformance_pack_ebs_common_tags = {
    service = "ebs"
  }
}

control "ebs_snapshot_not_publicly_restorable" {
  title       = "EBS snapshots should not be publicly restorable"
  description = "Manage access to the AWS Cloud by ensuring EBS snapshots are not publicly restorable."
  sql         = query.ebs_snapshot_not_publicly_restorable.sql

  tags = merge(local.conformance_pack_ebs_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "ebs_volume_encryption_at_rest_enabled" {
  title       = "EBS volume encryption at rest should be enabled"
  description = "Because sensitive data can exist and to help protect data at rest, ensure encryption is enabled for your Amazon Elastic Block Store (Amazon EBS) volumes."
  sql         = query.ebs_volume_encryption_at_rest_enabled.sql

  tags = merge(local.conformance_pack_ebs_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    rbi_cyber_security = "true"
  })
}

control "ebs_attached_volume_encryption_enabled" {
  title       = "Attached EBS volumes should have encryption enabled"
  description = "Because sensitive data can exist and to help protect data at rest, ensure encryption is enabled for your Amazon Elastic Block Store (Amazon EBS) volumes."
  sql         = query.ebs_attached_volume_encryption_enabled.sql

  tags = merge(local.conformance_pack_ebs_common_tags, {
    audit_manager_control_tower = "true"
    hipaa                       = "true"
    gdpr                        = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
    rbi_cyber_security          = "true"
  })
}

control "ebs_volume_in_backup_plan" {
  title       = "EBS volumes should be in a backup plan"
  description = "To help with data back-up processes, ensure your Amazon Elastic Block Store (Amazon EBS) volumes are a part of an AWS Backup plan."
  sql         = query.ebs_volume_in_backup_plan.sql

  tags = merge(local.conformance_pack_ebs_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "ebs_attached_volume_delete_on_termination_enabled" {
  title       = "Attached EBS volumes should have delete on termination enabled"
  description = "This rule ensures that Amazon Elastic Block Store volumes that are attached to Amazon Elastic Compute Cloud (Amazon EC2) instances are marked for deletion when an instance is terminated."
  sql         = query.ebs_attached_volume_delete_on_termination_enabled.sql

  tags = merge(local.conformance_pack_ebs_common_tags, {
    audit_manager_control_tower = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
  })
}

control "ebs_volume_protected_by_backup_plan" {
  title       = "EBS volumes should be protected by backup plan"
  description = "Ensure if Amazon Elastic Block Store (Amazon EBS) volumes are protected by a backup plan. The rule is non complaint if the Amazon EBS volume is not covered by a backup plan."
  sql         = query.ebs_volume_protected_by_backup_plan.sql

  tags = merge(local.conformance_pack_ebs_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}