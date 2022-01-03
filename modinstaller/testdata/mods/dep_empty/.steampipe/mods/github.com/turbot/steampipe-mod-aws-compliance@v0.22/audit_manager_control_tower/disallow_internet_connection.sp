locals {
  audit_manager_control_tower_disallow_internet_connection_common_tags = merge(local.audit_manager_control_tower_common_tags, {
    control_set = "disallow_internet_connection"
  })
}

benchmark "audit_manager_control_tower_disallow_internet_connection" {
  title         = "Disallow Internet Connection"
  description   = "This benchmark checks if the VPC security group restricts ingress from RDP and SSH."
  children = [
    benchmark.audit_manager_control_tower_disallow_internet_connection_2_0_1,
    benchmark.audit_manager_control_tower_disallow_internet_connection_2_0_2
  ]
  tags          = local.audit_manager_control_tower_disallow_internet_connection_common_tags
}

benchmark "audit_manager_control_tower_disallow_internet_connection_2_0_1" {
  title         = "2.0.1 - Disallow internet connection through RDP"
  description   = "Disallow internet connection through RDP - Checks whether security groups that are in use disallow unrestricted incoming TCP traffic to the specified"
  children = [
    control.vpc_security_group_restrict_ingress_common_ports_all
  ]

  tags = merge(local.audit_manager_control_tower_disallow_internet_connection_common_tags, {
    audit_manager_control_tower_item_id = "2.0.1"
  })
}

benchmark "audit_manager_control_tower_disallow_internet_connection_2_0_2" {
  title         = "2.0.2 - Disallow internet connection through SSH"
  description   = "Disallow internet connection through SSH - Checks whether security groups that are in use disallow unrestricted incoming SSH traffic."
  children = [
    control.vpc_security_group_restrict_ingress_ssh_all
  ]

  tags = merge(local.audit_manager_control_tower_disallow_internet_connection_common_tags, {
    audit_manager_control_tower_item_id = "2.0.2"
  })
}