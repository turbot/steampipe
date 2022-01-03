locals {
  conformance_pack_secretsmanager_common_tags = {
    service = "secretsmanager"
  }
}

control "secretsmanager_secret_automatic_rotation_enabled" {
  title       = "Secrets Manager secrets should have automatic rotation enabled"
  description = "This rule ensures AWS Secrets Manager secrets have rotation enabled. Rotating secrets on a regular schedule can shorten the period a secret is active, and potentially reduce the business impact if the secret is compromised."
  sql         = query.secretsmanager_secret_automatic_rotation_enabled.sql

  tags = merge(local.conformance_pack_secretsmanager_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
  })
}

control "secretsmanager_secret_rotated_as_scheduled" {
  title       = "Secrets Manager secrets should be rotated as per the rotation schedule"
  description = "This rule ensures that AWS Secrets Manager secrets have rotated successfully according to the rotation schedule. Rotating secrets on a regular schedule can shorten the period that a secret is active, and potentially reduce the business impact if it is compromised."
  sql         = query.secretsmanager_secret_rotated_as_scheduled.sql

  tags = merge(local.conformance_pack_secretsmanager_common_tags, {
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
  })
}
