locals {
  conformance_pack_iam_common_tags = {
    service = "iam"
  }
}

control "iam_account_password_policy_strong_min_reuse_24" {
  title       = "IAM password policies for users should have strong configurations"
  description = "The identities and the credentials are issued, managed, and verified based on an organizational IAM password policy."
  sql         = query.iam_account_password_policy_strong_min_reuse_24.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    rbi_cyber_security = "true"
  })
}

control "iam_group_not_empty" {
  title       = "IAM groups should have at least one user"
  description = "AWS Identity and Access Management (IAM) can help you incorporate the principles of least privilege and separation of duties with access permissions and authorizations, by ensuring that IAM groups have at least one IAM user."
  sql         = query.iam_group_not_empty.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
    soc_2             = "true"
  })
}

control "iam_policy_no_star_star" {
  title       = "IAM policy should not have statements with admin access"
  description = "AWS Identity and Access Management (IAM) can help you incorporate the principles of least privilege and separation of duties with access permissions and authorizations, restricting policies from containing 'Effect': 'Allow' with 'Action': '*' over 'Resource': '*'."
  sql         = query.iam_policy_no_star_star.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "iam_root_user_no_access_keys" {
  title       = "IAM root user should not have access keys"
  description = "Access to systems and assets can be controlled by checking that the root user does not have access keys attached to their AWS Identity and Access Management (IAM) role."
  sql         = query.iam_root_user_no_access_keys.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "iam_root_user_hardware_mfa_enabled" {
  title       = "IAM root user hardware MFA should be enabled"
  description = "Manage access to resources in the AWS Cloud by ensuring hardware MFA is enabled for the root user."
  sql         = query.iam_root_user_hardware_mfa_enabled.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr              = "true"
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
  })
}

control "iam_root_user_mfa_enabled" {
  title       = "IAM root user MFA should be enabled"
  description = "Manage access to resources in the AWS Cloud by ensuring MFA is enabled for the root user."
  sql         = query.iam_root_user_mfa_enabled.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    audit_manager_control_tower = "true"
    gdpr                        = "true"
    hipaa                       = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
  })
}

control "iam_user_access_key_age_90" {
  title       = "IAM user access keys should be rotated at least every 90 days"
  description = "The credentials are audited for authorized devices, users, and processes by ensuring IAM access keys are rotated as per organizational policy."
  sql         = query.iam_user_access_key_age_90.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr              = "true"
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
  })
}

control "iam_user_console_access_mfa_enabled" {
  title       = "IAM users with console access should have MFA enabled"
  description = "Manage access to resources in the AWS Cloud by ensuring that MFA is enabled for all AWS Identity and Access Management (IAM) users that have a console password."
  sql         = query.iam_user_console_access_mfa_enabled.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    audit_manager_control_tower = "true"
    gdpr                        = "true"
    hipaa                       = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
  })
}

control "iam_user_mfa_enabled" {
  title       = "IAM user MFA should be enabled"
  description = "Enable this rule to restrict access to resources in the AWS Cloud."
  sql         = query.iam_user_mfa_enabled.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    audit_manager_control_tower = "true"
    hipaa                       = "true"
    nist_800_53_rev_4           = "true"
    nist_csf                    = "true"
  })
}

control "iam_user_no_inline_attached_policies" {
  title       = "IAM user should not have any inline or attached policies"
  description = "This rule ensures AWS Identity and Access Management (IAM) policies are attached only to groups or roles to control access to systems and assets."
  sql         = query.iam_user_no_inline_attached_policies.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "iam_user_unused_credentials_90" {
  title       = "IAM user credentials that have not been used in 90 days should be disabled"
  description = "AWS Identity and Access Management (IAM) can help you with access permissions and authorizations by checking for IAM passwords and access keys that are not used for a specified time period."
  sql         = query.iam_user_unused_credentials_90.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr              = "true"
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
    soc_2             = "true"
  })
}

control "iam_user_in_group" {
  title       = "IAM users should be in at least one group"
  description = "AWS Identity and Access Management (IAM) can help you restrict access permissions and authorizations, by ensuring IAM users are members of at least one group."
  sql         = query.iam_user_in_group.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
    soc_2             = "true"
  })
}

control "iam_group_user_role_no_inline_policies" {
  title       = "IAM groups, users, and roles should not have any inline policies"
  description = "Ensure an AWS Identity and Access Management (IAM) user, IAM role or IAM group does not have an inline policy to control access to systems and assets."
  sql         = query.iam_group_user_role_no_inline_policies.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "iam_support_role" {
  title       = "Ensure a support role has been created to manage incidents with AWS Support"
  description = "AWS provides a support center that can be used for incident notification and response, as well as technical support and customer services."
  sql         = query.iam_support_role.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr = "true"
  })
}

control "iam_account_password_policy_min_length_14" {
  title       = "Ensure IAM password policy requires a minimum length of 14 or greater"
  description = "Password policies, in part, enforce password complexity requirements. Use IAM password policies to ensure that passwords are at least a given length. Security Hub recommends that the password policy require a minimum password length of 14 characters."
  sql         = query.iam_account_password_policy_min_length_14.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_account_password_policy_reuse_24" {
  title       = "Ensure IAM password policy prevents password reuse"
  description = "This control checks whether the number of passwords to remember is set to 24. The control fails if the value is not 24. IAM password policies can prevent the reuse of a given password by the same user."
  sql         = query.iam_account_password_policy_reuse_24.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_account_password_policy_strong" {
  title       = "Password policies for IAM users should have strong configurations"
  description = "This control checks whether the account password policy for IAM users have strong configurations."
  sql         = query.iam_account_password_policy_strong.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr = "true"
  })
}

control "iam_account_password_policy_one_lowercase_letter" {
  title       = "Ensure IAM password policy requires at least one lowercase letter"
  description = "Password policies, in part, enforce password complexity requirements. Use IAM password policies to ensure that passwords use different character sets. Security Hub recommends that the password policy require at least one lowercase letter. Setting a password complexity policy increases account resiliency against brute force login attempts."
  sql         = query.iam_account_password_policy_one_lowercase_letter.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_account_password_policy_one_uppercase_letter" {
  title       = "Ensure IAM password policy requires at least one uppercase letter"
  description = "Password policies, in part, enforce password complexity requirements. Use IAM password policies to ensure that passwords use different character sets."
  sql         = query.iam_account_password_policy_one_uppercase_letter.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_account_password_policy_one_number" {
  title       = "Ensure IAM password policy requires at least one number"
  description = "Password policies, in part, enforce password complexity requirements. Use IAM password policies to ensure that passwords use different character sets."
  sql         = query.iam_account_password_policy_one_number.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_password_policy_expire_90" {
  title       = "Ensure IAM password policy expires passwords within 90 days or less"
  description = "IAM password policies can require passwords to be rotated or expired after a given number of days. Security Hub recommends that the password policy expire passwords after 90 days or less. Reducing the password lifetime increases account resiliency against brute force login attempts."
  sql         = query.iam_password_policy_expire_90.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_account_password_policy_one_symbol" {
  title       = "Ensure IAM password policy requires at least one symbol"
  description = "Password policies, in part, enforce password complexity requirements. Use IAM password policies to ensure that passwords use different character sets. Security Hub recommends that the password policy require at least one symbol. Setting a password complexity policy increases account resiliency against brute force login attempts."
  sql         = query.iam_account_password_policy_one_symbol.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    gdpr  = "true"
    hipaa = "true"
  })
}

control "iam_all_policy_no_service_wild_card" {
  title       = "Ensure IAM policy should not grant full access to service"
  description = "Checks if AWS Identity and Access Management (IAM) policies grant permissions to all actions on individual AWS resources. The rule is non complaint if the managed IAM policy allows full access to at least 1 AWS service."
  sql         = query.iam_all_policy_no_service_wild_card.sql

  tags = merge(local.conformance_pack_iam_common_tags, {
    rbi_cyber_security = "true"
  })
}