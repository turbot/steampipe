benchmark "rbi_cyber_security_annex_i_7_3" {
  title       = "Annex I (7.3)"
  description = "Remote Desktop Protocol (RDP) which allows others to access the computer remotely over a network or over the internet should be always disabled and should be enabled only with the approval of the authorised officer of the UCB. Logs for such remote access shall be enabled and monitored for suspicious activities."

  children = [
    control.vpc_security_group_restrict_ingress_ssh_all
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_7_3"
  })
}
