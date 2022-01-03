benchmark "nist_800_53_rev_4_cm" {
  title       = "Configuration Management (CM)"
  description = "CM controls are specific to an organization’s configuration management policies. This includes a baseline configuration to operate as the basis for future builds or changes to information systems. Additionally, this includes information system component inventories and a security impact analysis control"
  children = [
    benchmark.nist_800_53_rev_4_cm_2,
    benchmark.nist_800_53_rev_4_cm_7,
    benchmark.nist_800_53_rev_4_cm_8
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_cm_2" {
  title       = "Baseline Configuration (CM-2)"
  description = "The organization develops, documents, and maintains under configuration control, a current baseline configuration of the information system."
  children = [
    control.cloudtrail_security_trail_enabled,
    control.ebs_attached_volume_delete_on_termination_enabled,
    control.ec2_instance_ssm_managed,
    control.ec2_stopped_instance_30_days,
    control.elb_application_lb_deletion_protection_enabled,
    control.ssm_managed_instance_compliance_association_compliant,
    control.vpc_security_group_restrict_ingress_common_ports_all
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_cm_7" {
  title       = "Least Functionality (CM-7)"
  description = "The organization configures the information system to provide only essential capabilities and prohibits or restricts the use of the functions, ports, protocols, and/or services."
  children = [
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_cm_8" {
  title       = "Information System Component Inventory (CM-8)"
  description = "The organization develops and documents an inventory of information system components that accurately reflects the current information system, includes all components within the authorization boundary of the information system, is at the level of granularity deemed necessary for tracking and reporting and reviews and updates the information system component inventory."
  children = [
    benchmark.nist_800_53_rev_4_cm_8_1,
    benchmark.nist_800_53_rev_4_cm_8_3
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_cm_8_1" {
  title       = "CM-8(1) Updates During Installation / Removals"
  description = "The organization updates the inventory of information system components as an integral part of component installations, removals, and information system updates."
  children = [
    control.ec2_instance_ssm_managed
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_cm_8_3" {
  title       = "CM-8(3) Automated Unauthorized Component Detection"
  description = "The organization employs automated mechanisms to detect the presence of unauthorized hardware, software, and firmware components within the information system and takes actions (disables network access by such components, isolates the components etc) when unauthorized components are detected."
  children = [
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant,
    control.ssm_managed_instance_compliance_patch_compliant
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
