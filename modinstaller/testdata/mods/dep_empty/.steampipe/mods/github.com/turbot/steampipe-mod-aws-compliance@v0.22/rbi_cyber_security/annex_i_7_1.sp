benchmark "rbi_cyber_security_annex_i_7_1" {
  title       = "Annex I (7.1)"
  description = "Disallow administrative rights on end-user workstations/PCs/laptops and provide access rights on a ‘need to know’ and ‘need to do’ basis."

  children = [
    control.iam_all_policy_no_service_wild_card,
    control.iam_group_user_role_no_inline_policies,
    control.iam_policy_no_star_star,
    control.iam_root_user_no_access_keys,
    control.iam_user_no_inline_attached_policies
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_7_1"
  })
}
