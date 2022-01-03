benchmark "rbi_cyber_security_annex_i_5_1" {
  title       = "Annex I (5.1)"
  description = "The firewall configurations should be set to the highest security level and evaluation of critical device (such as firewall, network switches, security devices, etc.) configurations should be done periodically."

  children = [
    control.apigateway_stage_use_waf_web_acl,
    control.elb_application_lb_waf_enabled,
    control.vpc_default_security_group_restricts_all_traffic,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_security_group_restrict_ingress_ssh_all,
    control.vpc_security_group_restrict_ingress_tcp_udp_all
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_5_1"
  })
}
