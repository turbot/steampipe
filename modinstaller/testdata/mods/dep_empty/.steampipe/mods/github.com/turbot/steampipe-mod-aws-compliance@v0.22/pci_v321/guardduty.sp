locals {
  pci_v321_guardduty_common_tags = merge(local.pci_v321_common_tags, {
    service = "guardduty"
  })
}

benchmark "pci_v321_guardduty" {
  title         = "GuardDuty"
  documentation = file("./pci_v321/docs/pci_v321_guardduty.md")
  children = [
    control.pci_v321_guardduty_1
  ]
  tags          = local.pci_v321_guardduty_common_tags
}

control "pci_v321_guardduty_1" {
  title         = "1 GuardDuty should be enabled"
  description   = "This control checks whether Amazon GuardDuty is enabled in your AWS account and Region."
  severity      = "high"
  sql           = query.guardduty_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_guardduty_1.md")

  tags = merge(local.pci_v321_guardduty_common_tags, {
    pci_item_id      = "guardduty_1"
    pci_requirements = "11.4"
  })
}