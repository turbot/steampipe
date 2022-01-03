locals {
  foundational_security_dms_common_tags = merge(local.foundational_security_common_tags, {
    service = "dms"
  })
}

benchmark "foundational_security_dms" {
  title         = "DMS"
  documentation = file("./foundational_security/docs/foundational_security_dms.md")
  children = [
    control.foundational_security_dms_1
  ]
  tags          = local.foundational_security_dms_common_tags
}

control "foundational_security_dms_1" {
  title         = "1 AWS Database Migration Service replication instances should not be public"
  description   = "This control checks whether AWS DMS replication instances are public. To do this, it examines the value of the PubliclyAccessible field. A private replication instance has a private IP address that you cannot access outside of the replication network. A replication instance should have a private IP address when the source and target databases are in the same network. The network must also be connected to the replication instance's VPC using a VPN, AWS Direct Connect, or VPC peering. To learn more about public and private replication instances, see Public and private replication instances in the AWS Database Migration Service User Guide."
  severity      = "critical"
  sql           = query.dms_replication_instance_not_publicly_accessible.sql
  documentation = file("./foundational_security/docs/foundational_security_dms_1.md")

  tags = merge(local.foundational_security_dms_common_tags, {
    foundational_security_item_id  = "dms_1"
    foundational_security_category = "secure_network_configuration"
  })
}