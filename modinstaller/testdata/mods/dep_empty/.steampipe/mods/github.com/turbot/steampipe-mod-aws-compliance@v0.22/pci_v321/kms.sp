locals {
  pci_v321_kms_common_tags = merge(local.pci_v321_common_tags, {
    service = "kms"
  })
}

benchmark "pci_v321_kms" {
  title         = "KMS"
  documentation = file("./pci_v321/docs/pci_v321_kms.md")
  children = [
    control.pci_v321_kms_1
  ]
  tags          = local.pci_v321_kms_common_tags
}

control "pci_v321_kms_1" {
  title         = "1 Customer master key (CMK) rotation should be enabled"
  description   = "This control checks that key rotation is enabled for each customer master key (CMK). It does not check CMKs that have imported key material. You should ensure keys that have imported material and those that are not stored in AWS KMS are rotated. AWS managed customer master keys are rotated once every 3 years."
  severity      = "medium"
  sql           = query.kms_cmk_rotation_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_kms_1.md")

  tags = merge(local.pci_v321_kms_common_tags, {
    pci_item_id      = "kms_1"
    pci_requirements = "3.6.4"
  })
}