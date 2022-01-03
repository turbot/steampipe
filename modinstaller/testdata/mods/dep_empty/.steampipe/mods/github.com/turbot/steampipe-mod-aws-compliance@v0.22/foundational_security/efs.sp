locals {
  foundational_security_efs_common_tags = merge(local.foundational_security_common_tags, {
    service = "efs"
  })
}

benchmark "foundational_security_efs" {
  title         = "EFS"
  documentation = file("./foundational_security/docs/foundational_security_efs.md")
  children = [
    control.foundational_security_efs_1,
    control.foundational_security_efs_2
  ]
  tags          = local.foundational_security_efs_common_tags
}

control "foundational_security_efs_1" {
  title         = "1 Amazon EFS should be configured to encrypt file data at rest using AWS KMS"
  description   = "This control checks whether Amazon Elastic File System is configured to encrypt the file data using AWS KMS."
  severity      = "medium"
  sql           = query.efs_file_system_encrypt_data_at_rest.sql
  documentation = file("./foundational_security/docs/foundational_security_efs_1.md")

  tags = merge(local.foundational_security_efs_common_tags, {
    foundational_security_item_id  = "efs_1"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}

control "foundational_security_efs_2" {
  title         = "2 Amazon EFS volumes should be in backup plans"
  description   = "This control checks whether Amazon Elastic File System (Amazon EFS) file systems are added to the backup plans in AWS Backup. The control fails if Amazon EFS file systems are not included in the backup plans."
  severity      = "medium"
  sql           = query.efs_file_system_automatic_backups_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_efs_2.md")

  tags = merge(local.foundational_security_efs_common_tags, {
    foundational_security_item_id  = "efs_2"
    foundational_security_category = "backup"
  })
}