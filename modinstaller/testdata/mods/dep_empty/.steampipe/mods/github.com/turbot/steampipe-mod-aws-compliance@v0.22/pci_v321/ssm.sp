locals {
  pci_v321_ssm_common_tags = merge(local.pci_v321_common_tags, {
    service = "ssm"
  })
}

benchmark "pci_v321_ssm" {
  title         = "SSM"
  documentation = file("./pci_v321/docs/pci_v321_ssm.md")
  children = [
    control.pci_v321_ssm_1,
    control.pci_v321_ssm_2,
    control.pci_v321_ssm_3
  ]
  tags = local.pci_v321_ssm_common_tags
}

control "pci_v321_ssm_1" {
  title         = "1 Amazon EC2 instances managed by Systems Manager should have a patch compliance status of COMPLIANT after a patch installation"
  description   = "This control checks whether the compliance status of the Amazon EC2 Systems Manager patch compliance is COMPLIANT or NON_COMPLIANT after the patch installation on the instance."
  severity      = "medium"
  sql           = query.ssm_managed_instance_compliance_patch_compliant.sql
  documentation = file("./pci_v321/docs/pci_v321_ssm_1.md")

  tags = merge(local.pci_v321_ssm_common_tags, {
    pci_item_id      = "ssm_1"
    pci_requirements = "6.2"
  })
}

control "pci_v321_ssm_2" {
  title         = "2 Instances managed by Systems Manager should have an association compliance status of COMPLIANT"
  description   = "This control checks whether the status of the AWS Systems Manager association compliance is COMPLIANT or NON_COMPLIANT after the association is run on an instance. The control passes if the association compliance status is COMPLIANT."
  severity      = "low"
  sql           = query.ssm_managed_instance_compliance_association_compliant.sql
  documentation = file("./pci_v321/docs/pci_v321_ssm_2.md")

  tags = merge(local.pci_v321_ssm_common_tags, {
    pci_item_id      = "ssm_2"
    pci_requirements = "2.4"
  })
}

control "pci_v321_ssm_3" {
  title         = "3 EC2 instances should be managed by AWS Systems Manager"
  description   = "This control checks whether the EC2 instances in your account are managed by Systems Manager."
  severity      = "medium"
  sql           = query.ec2_instance_ssm_managed.sql
  documentation = file("./pci_v321/docs/pci_v321_ssm_3.md")

  tags = merge(local.pci_v321_ssm_common_tags, {
    pci_item_id      = "ssm_3"
    pci_requirements = "2.4,6.2"
  })
}