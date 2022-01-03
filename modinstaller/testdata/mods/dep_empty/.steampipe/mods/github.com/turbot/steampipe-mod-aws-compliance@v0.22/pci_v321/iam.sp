locals {
  pci_v321_iam_common_tags = merge(local.pci_v321_common_tags, {
    service = "iam"
  })
}

benchmark "pci_v321_iam" {
  title         = "IAM"
  documentation = file("./pci_v321/docs/pci_v321_iam.md")
  children = [
    control.pci_v321_iam_1,
    control.pci_v321_iam_2,
    control.pci_v321_iam_3,
    control.pci_v321_iam_4,
    control.pci_v321_iam_5,
    control.pci_v321_iam_6,
    control.pci_v321_iam_7,
    control.pci_v321_iam_8,
  ]
  tags = local.pci_v321_iam_common_tags
}

control "pci_v321_iam_1" {
  title         = "1 IAM root user access key should not exist"
  description   = "This control checks whether user access keys exist for the root user."
  severity      = "critical"
  sql           = query.iam_root_user_no_access_keys.sql
  documentation = file("./pci_v321/docs/pci_v321_iam_1.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_1"
    pci_requirements = "2.1,2.2,7.2.1"
  })
}

control "pci_v321_iam_2" {
  title         = "2 IAM users should not have IAM policies attached"
  description   = "This control checks that none of your IAM users have policies attached. IAM users must inherit permissions from IAM groups or roles. It does not check whether least privileged policies are applied to IAM roles and groups."
  severity      = "low"
  sql           = query.iam_user_no_inline_attached_policies.sql
  documentation = file("./pci_v321/docs/pci_v321_iam_2.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_2"
    pci_requirements = "7.2.1"
  })
}

control "pci_v321_iam_3" {
  title         = "3 IAM policies should not allow full '*' administrative privileges"
  description   = "This control checks whether the default version of AWS Identity and Access Management policies (also known as customer managed policies) do not have administrator access with a statement that has 'Effect': 'Allow' with 'Action': '*' over 'Resource': '*'."
  severity      = "high"
  sql           = query.iam_policy_no_star_star.sql
  documentation = file("./pci_v321/docs/pci_v321_iam_3.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_3"
    pci_requirements = "7.2.1"
  })
}

control "pci_v321_iam_4" {
  title         = "4 Hardware MFA should be enabled for the root user"
  description   = "This control checks whether your AWS account is enabled to use multi-factor authentication (MFA) hardware device to sign in with root user credentials. It does not check whether you are using virtual MFA."
  severity      = "critical"
  sql           = query.iam_root_user_hardware_mfa_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_iam_4.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_4"
    pci_requirements = "8.3.1"
  })
}

control "pci_v321_iam_5" {
  title         = "5 Virtual MFA should be enabled for the root user"
  description   = "This control checks whether users of your AWS account require a multi-factor authentication (MFA) device to sign in with root user credentials. It does not check whether you are using hardware MFA."
  severity      = "critical"
  sql           = query.iam_root_user_virtual_mfa.sql
  #documentation = file("./pci_v321/docs/pci_v321_iam_5.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_5"
    pci_requirements = "8.3.1"
  })
}

control "pci_v321_iam_6" {
  title         = "6 MFA should be enabled for all IAM users"
  description   = "This control checks whether the IAM users have multi-factor authentication (MFA) enabled."
  severity      = "medium"
  sql           = query.iam_user_mfa_enabled.sql
  #documentation = file("./pci_v321/docs/pci_v321_iam_6.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_6"
    pci_requirements = "8.3.1"
  })
}

control "pci_v321_iam_7" {
  title         = "7 IAM user credentials should be disabled if not used within a predefined number of days"
  description   = "This control checks whether your IAM users have passwords or active access keys that have not been used within a specified number of days. The default is 90 days."
  severity      = "medium"
  sql           = query.iam_user_unused_credentials_90.sql
  #documentation = file("./pci_v321/docs/pci_v321_iam_7.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_7"
    pci_requirements = "8.1.4"
  })
}

control "pci_v321_iam_8" {
  title         = "8 Password policies for IAM users should have strong configurations"
  description   = "This control checks whether the account password policy for IAM users uses the following minimum PCI DSS configurations."
  severity      = "medium"
  sql           = query.iam_account_password_policy_strong.sql
  #documentation = file("./pci_v321/docs/pci_v321_iam_8.md")

  tags = merge(local.pci_v321_iam_common_tags, {
    pci_item_id      = "iam_8"
    pci_requirements = "8.1.4,8.2.3,8.2.4,8.2.5"
  })
}