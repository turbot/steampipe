benchmark "rbi_cyber_security_annex_i_7_2" {
  title       = "Annex I (7.2)"
  description = "Passwords should be set as complex and lengthy and users should not use same passwords for all the applications/systems/devices."

  children = [
    control.iam_account_password_policy_strong_min_reuse_24
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_7_2"
  })
}
