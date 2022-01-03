locals {
  pci_v321_config_common_tags = merge(local.pci_v321_common_tags, {
    service = "config"
  })
}

benchmark "pci_v321_config" {
  title         = "Config"
  documentation = file("./pci_v321/docs/pci_v321_config.md")
  children = [
    control.pci_v321_config_1
  ]
  tags = local.pci_v321_config_common_tags
}

control "pci_v321_config_1" {
  title         = "1 AWS Config should be enabled"
  description   = "This control checks whether AWS Config is enabled in the account for the local Region and is recording all resources. It does not check for change detection for all critical system files and content files, as AWS Config supports only a subset of resource types. The AWS Config service performs configuration management of supported AWS resources in your account and delivers log files to you. The recorded information includes the configuration item (AWS resource), relationships between configuration items, and any configuration changes between resources."
  severity      = "medium"
  sql           = query.config_enabled_all_regions.sql
  documentation = file("./pci_v321/docs/pci_v321_config_1.md")

  tags = merge(local.pci_v321_config_common_tags, {
    pci_item_id      = "config_1"
    pci_requirements = "10.5.2,11.5"
  })
}