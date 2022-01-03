benchmark "nist_800_53_rev_4_ia" {
  title       = "Identification and Authentication (IA)"
  description = "IA controls are specific to the identification and authentication policies in an organization. This includes the identification and authentication of organizational and non-organizational users and how the management of those systems."
  children = [
    benchmark.nist_800_53_rev_4_ia_2,
    benchmark.nist_800_53_rev_4_ia_5
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_2" {
  title       = "Identification and Authentication (Organizational users) (IA-2)"
  description = "The information system uniquely identifies and authenticates organizational users (or processes acting on behalf of organizational users)."
  children = [
    benchmark.nist_800_53_rev_4_ia_2_1,
    benchmark.nist_800_53_rev_4_ia_2_2,
    benchmark.nist_800_53_rev_4_ia_2_11,
    control.iam_account_password_policy_strong_min_reuse_24
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_2_1" {
  title       = "IA-2(1) Network Access To Privileged Accounts"
  description = "The information system implements multi-factor authentication for network access to privileged accounts."
  children = [
    control.iam_root_user_hardware_mfa_enabled,
    control.iam_root_user_mfa_enabled,
    control.iam_user_console_access_mfa_enabled,
    control.iam_user_mfa_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_2_2" {
  title       = "IA-2(2) Network Access To Non-Privileged Accounts"
  description = "The information system implements multifactor authentication for network access to non-privileged accounts."
  children = [
    control.iam_user_console_access_mfa_enabled,
    control.iam_user_mfa_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_2_11" {
  title       = "IA-2(11) Remote Access - Separate Device"
  description = "The information system implements multifactor authentication for remote access to privileged and non-privileged accounts such that one of the factors is provided by a device separate from the system gaining access and the device meets [Assignment: organization-defined strength of mechanism requirements]."
  children = [
    control.iam_root_user_hardware_mfa_enabled,
    control.iam_root_user_mfa_enabled,
    control.iam_user_console_access_mfa_enabled,
    control.iam_user_mfa_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_5" {
  title       = "Authenticator Management (IA-5)"
  description = "Authenticate users and devices. Automate administrative control. Enforce restrictions. Protect against unauthorized use."
  children = [
    benchmark.nist_800_53_rev_4_ia_5_1,
    benchmark.nist_800_53_rev_4_ia_5_4,
    benchmark.nist_800_53_rev_4_ia_5_7
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_5_1" {
  title       = "IA-5(1) Password-Based Authentication"
  description = "The information system, for password-based authentication that enforces  minimum password complexity, stores and transmits only cryptographically-protected passwords, enforces password minimum and maximum lifetime restrictions, prohibits password reuse, allows the use of a temporary password for system logons with an immediate change to a permanent password etc."
  children = [
    control.iam_account_password_policy_strong_min_reuse_24
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_5_4" {
  title       = "IA-5(4) Automated Support For Password Strength Determination"
  description = "The organization employs automated tools to determine if password authenticators are sufficiently strong to satisfy [Assignment: organization-defined requirements]."
  children = [
    control.iam_account_password_policy_strong_min_reuse_24
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ia_5_7" {
  title       = "IA-5(7) No Embedded Unencrypted Static Authenticators"
  description = "The organization ensures that unencrypted static authenticators are not embedded in applications or access scripts or stored on function keys."
  children = [
    control.codebuild_project_plaintext_env_variables_no_sensitive_aws_values
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
