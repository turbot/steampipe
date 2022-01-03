locals {
  foundational_security_iam_common_tags = merge(local.foundational_security_common_tags, {
    service = "iam"
  })
}

benchmark "foundational_security_iam" {
  title         = "IAM"
  documentation = file("./foundational_security/docs/foundational_security_iam.md")
  children = [
    control.foundational_security_iam_1,
    control.foundational_security_iam_2,
    control.foundational_security_iam_3,
    control.foundational_security_iam_4,
    control.foundational_security_iam_5,
    control.foundational_security_iam_6,
    control.foundational_security_iam_7,
    control.foundational_security_iam_8,
    control.foundational_security_iam_21
  ]
  tags          = local.foundational_security_iam_common_tags
}

control "foundational_security_iam_1" {
  title         = "1 IAM policies should not allow full '*' administrative privileges"
  description   = "This control checks whether the default version of IAM policies (also known as customer managed policies) has administrator access that includes a statement with 'Effect': 'Allow' with 'Action': '*' over 'Resource': '*'. The control only checks the customer managed policies that you create. It does not check inline and AWS managed policies."
  severity      = "high"
  sql           = query.iam_custom_policy_no_star_star.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_1.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_1"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_2" {
  title         = "2 IAM users should not have IAM policies attached"
  description   = "This control checks that none of your IAM users have policies attached. Instead, IAM users must inherit permissions from IAM groups or roles."
  severity      = "low"
  sql           = query.iam_user_no_inline_attached_policies.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_2.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_2"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_3" {
  title         = "3 IAM users' access keys should be rotated every 90 days or less"
  description   = "This control checks whether the active access keys are rotated within 90 days."
  severity      = "medium"
  sql           = query.iam_user_access_key_age_90.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_3.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_3"
    #foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_4" {
  title         = "4 IAM root user access key should not exist"
  description   = "This control checks whether the root user access key is present. The root account is the most privileged user in an AWS account. AWS access keys provide programmatic access to a given account."
  severity      = "critical"
  sql           = query.iam_root_user_no_access_keys.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_4.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_4"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_5" {
  title         = "5 MFA should be enabled for all IAM users that have a console password"
  description   = "This control checks whether AWS multi-factor authentication (MFA) is enabled for all IAM users that use a console password."
  severity      = "medium"
  sql           = query.iam_user_console_access_mfa_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_5.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_5"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_6" {
  title         = "6 Hardware MFA should be enabled for the root user"
  description   = "This control checks whether your AWS account is enabled to use a hardware multi-factor authentication (MFA) device to sign in with root user credentials."
  severity      = "critical"
  sql           = query.iam_root_user_hardware_mfa_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_6.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_6"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_7" {
  title         = "7 Password policies for IAM users should have strong configurations"
  description   = "This control checks whether the account password policy for IAM users uses the recommended configurations."
  severity      = "medium"
  sql           = query.iam_account_password_policy_strong_min_length_8.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_7.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_7"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_8" {
  title         = "8 Unused IAM user credentials should be removed"
  description   = "This control checks whether your IAM users have passwords or active access keys that have not been used for 90 days."
  severity      = "medium"
  sql           = query.iam_user_unused_credentials_90.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_8.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_8"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_iam_21" {
  title         = "21 IAM customer managed policies that you create should not allow wildcard actions for services"
  description   = "This control checks whether the IAM identity-based policies that you create have Allow statements that use the * wildcard to grant permissions for all actions on any service. The control fails if any policy statement includes 'Effect': 'Allow' with 'Action': 'Service:*'."
  severity      = "low"
  sql           = query.iam_custom_policy_no_service_wild_card.sql
  documentation = file("./foundational_security/docs/foundational_security_iam_21.md")

  tags = merge(local.foundational_security_iam_common_tags, {
    foundational_security_item_id  = "iam_21"
    foundational_security_category = "secure_access_management"
  })
}