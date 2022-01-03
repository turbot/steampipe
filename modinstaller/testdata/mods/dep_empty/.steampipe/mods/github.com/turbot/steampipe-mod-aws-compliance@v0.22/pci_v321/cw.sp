locals {
  pci_v321_cw_common_tags = merge(local.pci_v321_common_tags, {
    service = "cloudwatch"
  })
}

benchmark "pci_v321_cw" {
  title         = "CloudWatch"
  documentation = file("./pci_v321/docs/pci_v321_cw.md")
  children = [
    control.pci_v321_cw_1
  ]
  tags          = local.pci_v321_cw_common_tags
}

control "pci_v321_cw_1" {
  title         = "1 A log metric filter and alarm should exist for usage of the 'root' user"
  description   = "This control checks for the CloudWatch metric filters using the following pattern: { $.userIdentity.type = 'Root' && $.userIdentity.invokedBy NOT EXISTS && $.eventType != AwsServiceEvent }."
  severity      = "critical"
  sql           = query.log_metric_filter_root_login.sql
  documentation = file("./pci_v321/docs/pci_v321_cw_1.md")

  tags = merge(local.pci_v321_cw_common_tags, {
    pci_item_id      = "cw_1"
    pci_requirements = "7.2.1"
  })
}
