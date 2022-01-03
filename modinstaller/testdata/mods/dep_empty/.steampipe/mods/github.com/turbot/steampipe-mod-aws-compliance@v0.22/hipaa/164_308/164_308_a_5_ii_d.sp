benchmark "hipaa_164_308_a_5_ii_d" {
  title       = "164.308(a)(5)(ii)(D) Password management"
  description = "Procedures for creating, changing, and safeguarding passwords."
  children = [
    control.iam_account_password_policy_min_length_14,
    control.iam_account_password_policy_one_lowercase_letter,
    control.iam_account_password_policy_one_number,
    control.iam_account_password_policy_one_symbol,
    control.iam_account_password_policy_one_uppercase_letter,
    control.iam_account_password_policy_reuse_24,
    control.iam_password_policy_expire_90,
    control.iam_user_access_key_age_90,
    control.iam_user_unused_credentials_90
  ]

  tags = merge(local.hipaa_164_308_common_tags, {
    hipaa_item_id = "164_308_a_5_ii_d"
  })
}