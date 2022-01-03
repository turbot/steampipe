benchmark "hipaa_164_312_d" {
  title          = "164.312(d) Person or entity authentication"
  description    = "Implement procedures to verify that a person or entity seeking access to electronic protected health information is the one claimed."
  children = [
    control.iam_account_password_policy_strong_min_reuse_24,
    control.iam_root_user_hardware_mfa_enabled,
    control.iam_root_user_mfa_enabled,
    control.iam_user_console_access_mfa_enabled,
    control.iam_user_mfa_enabled
  ]

  tags = merge(local.hipaa_164_312_common_tags, {
    hipaa_item_id = "164_312_d"
  })
}