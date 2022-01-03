locals {
  pci_v321_dms_common_tags = merge(local.pci_v321_common_tags, {
    service = "dms"
  })
}

benchmark "pci_v321_dms" {
  title         = "DMS"
  documentation = file("./pci_v321/docs/pci_v321_dms.md")
  children = [
    control.pci_v321_dms_1
  ]
  tags = local.pci_v321_dms_common_tags
}

control "pci_v321_dms_1" {
  title         = "1 AWS Database Migration Service replication instances should not be public"
  description   = "This control checks whether AWS DMS replication instances are public. To do this, it examines the value of the PubliclyAccessible field. A private replication instance has a private IP address that you cannot access outside of the replication network. A replication instance should have a private IP address when the source and target databases are in the same network, and the network is connected to the replication instance's VPC using a VPN, AWS Direct Connect, or VPC peering."
  severity      = "critical"
  sql           = query.dms_replication_instance_not_publicly_accessible.sql
  documentation = file("./pci_v321/docs/pci_v321_dms_1.md")

  tags = merge(local.pci_v321_dms_common_tags, {
    pci_item_id      = "dms_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6"
  })
}