locals {
  pci_v321_rds_common_tags = merge(local.pci_v321_common_tags, {
    service = "rds"
  })
}

benchmark "pci_v321_rds" {
  title         = "RDS"
  documentation = file("./pci_v321/docs/pci_v321_rds.md")
  children = [
    control.pci_v321_rds_1,
    control.pci_v321_rds_2,
  ]
  tags          = local.pci_v321_rds_common_tags
}

control "pci_v321_rds_1" {
  title         = "1 RDS snapshots should prohibit public access"
  description   = "This control checks whether Amazon RDS DB snapshots prohibit access by other accounts. You should also ensure that access to the snapshot and permission to change Amazon RDS configuration is restricted to authorized principals only."
  severity      = "critical"
  sql           = query.rds_db_snapshot_prohibit_public_access.sql
  documentation = file("./pci_v321/docs/pci_v321_rds_1.md")

  tags = merge(local.pci_v321_rds_common_tags, {
    pci_item_id      = "rds_1"
    pci_requirements = "1.2.1,1.3.1,1.3.4,1.3.6,7.2.1"
  })
}

control "pci_v321_rds_2" {
  title         = "2 RDS DB Instances should prohibit public access"
  description   = "This control checks whether RDS instances are publicly accessible by evaluating the publiclyAccessible field in the instance configuration item. The value of publiclyAccessible indicates whether the DB instance is publicly accessible. When the DB instance is publicly accessible,it is an Internet-facing instance with a publicly resolvable DNS name, which resolves to a public IP address. When the DB instance isn't publicly accessible, it is an internal instance with a DNS name that resolves to a private IP address."
  severity      = "critical"
  sql           = query.rds_db_instance_prohibit_public_access.sql
  documentation = file("./pci_v321/docs/pci_v321_rds_2.md")

  tags = merge(local.pci_v321_rds_common_tags, {
    pci_item_id      = "rds_2"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6,7.2.1"
  })
}