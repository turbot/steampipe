locals {
  conformance_pack_kms_common_tags = {
    service = "kms"
  }
}

control "kms_key_not_pending_deletion" {
  title       = "KMS keys should not be pending deletion"
  description = "To help protect data at rest, ensure necessary customer master keys (CMKs) are not scheduled for deletion in AWS Key Management Service (AWS KMS)."
  sql         = query.kms_key_not_pending_deletion.sql

  tags = merge(local.conformance_pack_kms_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "kms_cmk_rotation_enabled" {
  title       = "KMS CMK rotation should be enabled"
  description = "Enable key rotation to ensure that keys are rotated once they have reached the end of their crypto period."
  sql         = query.kms_cmk_rotation_enabled.sql

  tags = merge(local.conformance_pack_kms_common_tags, {
    hippa              = "true"
    gdpr               = "true"
    nist_800_53_rev_4  = "true"
    rbi_cyber_security = "true"
  })
}

control "kms_key_decryption_restricted_in_iam_customer_managed_policy" {
  title      = "KMS key decryption should be restricted in IAM customer managed policy"
  description = "Checks whether the default version of IAM customer managed policies allow principals to use the AWS KMS decryption actions on all resources. This control uses Zelkova, an automated reasoning engine, to validate and warn you about policies that may grant broad access to your secrets across AWS accounts. This control fails if the kms:Decrypt or kms:ReEncryptFrom actions are allowed on all KMS keys. The control evaluates both attached and unattached customer managed policies. It does not check inline policies or AWS managed policies."
  sql         = query.kms_key_decryption_restricted_in_iam_customer_managed_policy.sql

  tags = merge(local.conformance_pack_kms_common_tags, {
    hipaa = "true"
  })
}

control "kms_key_decryption_restricted_in_iam_inline_policy" {
  title       = "KMS key decryption should be restricted in IAM inline policy"
  description = "Checks whether the inline policies that are embedded in your IAM identities (role, user, or group) allow the AWS KMS decryption actions on all KMS keys. This control uses Zelkova, an automated reasoning engine, to validate and warn you about policies that may grant broad access to your secrets across AWS accounts. This control fails if kms:Decrypt or kms:ReEncryptFrom actions are allowed on all KMS keys in an inline policy."
  sql         = query.kms_key_decryption_restricted_in_iam_inline_policy.sql

  tags = merge(local.conformance_pack_kms_common_tags, {
    hipaa = "true"
  })
}

