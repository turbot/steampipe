locals {
  audit_manager_control_tower_ebs_checks_common_tags = merge(local.audit_manager_control_tower_common_tags, {
    control_set = "ebs_checks"
  })
}

benchmark "audit_manager_control_tower_ebs_checks" {
  title         = "EBS checks"
  description   = "This benchmark checks if EBS volumes are in use, encrypted etc."
  children = [
    benchmark.audit_manager_control_tower_ebs_checks_1_0_1,
    benchmark.audit_manager_control_tower_ebs_checks_1_0_2,
    benchmark.audit_manager_control_tower_ebs_checks_1_0_3
  ]
  tags          = local.audit_manager_control_tower_ebs_checks_common_tags
}

benchmark "audit_manager_control_tower_ebs_checks_1_0_1" {
  title         = "1.0.1 - Disallow launch of EC2 instance types that are not EBS-optimized"
  description   = "Disallow launch of EC2 instance types that are not EBS-optimized - Checks whether EBS optimization is enabled for your EC2 instances that can be EBS-optimized."
  children = [
    control.ec2_instance_ebs_optimized
  ]

  tags = merge(local.audit_manager_control_tower_ebs_checks_common_tags, {
    audit_manager_control_tower_item_id = "1.0.1"
  })
}

benchmark "audit_manager_control_tower_ebs_checks_1_0_2" {
  title         = "1.0.2 - Disallow EBS volumes that are unattached to an EC2 instance"
  description   = "Disallow EBS volumes that are unattached to an EC2 instance - Checks whether EBS volumes are attached to EC2 instances"
  children = [
    control.ebs_attached_volume_delete_on_termination_enabled
  ]

  tags = merge(local.audit_manager_control_tower_ebs_checks_common_tags, {
    audit_manager_control_tower_item_id = "1.0.2"
  })
}

benchmark "audit_manager_control_tower_ebs_checks_1_0_3" {
  title         = "1.0.3 - Enable encryption for EBS volumes attached to EC2 instances"
  description   = "Enable encryption for EBS volumes attached to EC2 instances - Checks whether EBS volumes that are in an attached state are encrypted."
  children = [
    control.ebs_attached_volume_encryption_enabled
  ]

  tags = merge(local.audit_manager_control_tower_ebs_checks_common_tags, {
    audit_manager_control_tower_item_id = "1.0.3"
  })
}