locals {
  conformance_pack_backup_common_tags = {
    service = "backup"
  }
}

control "backup_recovery_point_manual_deletion_disabled" {
  title       = "Backup recovery point manual deletion should be disabled"
  description = "Checks if a backup vault has an attached resource-based policy which prevents deletion of recovery points. The rule is non complaint if the Backup Vault does not have resource-based policies or has policies without a suitable 'Deny' statement."
  sql         = query.backup_recovery_point_manual_deletion_disabled.sql

  tags = merge(local.conformance_pack_backup_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}

control "backup_plan_min_retention_35_days" {
  title       = "Backup plan min frequency and min retention check"
  description = "Checks if a backup plan has a backup rule that satisfies the required frequency and retention period(35 Days). The rule is non complaint if recovery points are not created at least as often as the specified frequency or expire before the specified period."
  sql         = query.backup_plan_min_retention_35_days.sql

  tags = merge(local.conformance_pack_backup_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}

control "backup_recovery_point_encryption_enabled" {
  title       = "Backup recovery point should be encrypted"
  description = "Ensure if a recovery point is encrypted. The rule is non complaint if the recovery point is not encrypted."
  sql         = query.backup_recovery_point_encryption_enabled.sql

  tags = merge(local.conformance_pack_backup_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}