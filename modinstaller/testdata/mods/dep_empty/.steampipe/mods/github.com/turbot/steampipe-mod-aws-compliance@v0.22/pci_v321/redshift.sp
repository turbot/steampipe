locals {
  pci_v321_redshift_common_tags = merge(local.pci_v321_common_tags, {
    service = "redshift"
  })
}

benchmark "pci_v321_redshift" {
  title         = "Redshift"
  documentation = file("./pci_v321/docs/pci_v321_redshift.md")
  children = [
    control.pci_v321_redshift_1
  ]
  tags          = local.pci_v321_redshift_common_tags
}

control "pci_v321_redshift_1" {
  title         = "1 Amazon Redshift clusters should prohibit public access"
  description   = "This control checks whether Amazon Redshift clusters are publicly accessible by evaluating the publiclyAccessible field in the cluster configuration item."
  severity      = "critical"
  sql           = query.redshift_cluster_prohibit_public_access.sql
  documentation = file("./pci_v321/docs/pci_v321_redshift_1.md")

  tags = merge(local.pci_v321_redshift_common_tags, {
    pci_item_id      = "redshift_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6"
  })
}