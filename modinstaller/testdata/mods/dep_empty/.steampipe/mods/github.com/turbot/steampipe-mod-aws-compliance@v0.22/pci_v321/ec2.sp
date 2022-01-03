locals {
  pci_v321_ec2_common_tags = merge(local.pci_v321_common_tags, {
    service = "ec2"
  })
}

benchmark "pci_v321_ec2" {
  title         = "EC2"
  documentation = file("./pci_v321/docs/pci_v321_ec2.md")
  children = [
    control.pci_v321_ec2_1,
    control.pci_v321_ec2_2,
    control.pci_v321_ec2_3,
    control.pci_v321_ec2_4,
    control.pci_v321_ec2_5,
    control.pci_v321_ec2_6,
  ]
  tags = local.pci_v321_ec2_common_tags
}

control "pci_v321_ec2_1" {
  title         = "1 Amazon EBS snapshots should not be publicly restorable"
  description   = "This control checks whether Amazon Elastic Block Store snapshots are not publicly restorable by everyone, which makes them public. Amazon EBS snapshots should not be publicly restorable by everyone unless you explicitly allow it, to avoid accidental exposure of your company’s sensitive data."
  severity      = "critical"
  sql           = query.ebs_snapshot_not_publicly_restorable.sql
  documentation = file("./pci_v321/docs/pci_v321_ec2_1.md")

  tags = merge(local.pci_v321_ec2_common_tags, {
    pci_item_id      = "ec2_1"
    pci_requirements = "1.2.1,1.3.1,1.3.4,1.3.4,7.2.1"
  })
}

control "pci_v321_ec2_2" {
  title         = "2 VPC default security group should prohibit inbound and outbound traffic"
  description   = "This control checks that the default security group of a VPC does not allow inbound or outbound traffic. It does not check for access restrictions for other security groups that are not default, and other VPC configurations."
  severity      = "medium"
  sql           = query.vpc_default_security_group_restricts_all_traffic.sql
  documentation = file("./pci_v321/docs/pci_v321_ec2_2.md")

  tags = merge(local.pci_v321_ec2_common_tags, {
    pci_item_id      = "ec2_2"
    pci_requirements = "1.2.1,1.3.4,2.1"
  })
}

control "pci_v321_ec2_3" {
  title         = "3 Unused EC2 security groups should be removed"
  description   = "This control helps you maintain an accurate asset inventory of needed security groups in your cardholder data environment (CDE). It does so by checking that security groups are attached to Amazon EC2 instances or to an ENI. A failed finding indicates you may have unused Amazon EC2 security groups."
  severity      = "low"
  sql           = query.vpc_security_group_associated.sql
  documentation = file("./pci_v321/docs/pci_v321_ec2_3.md")

  tags = merge(local.pci_v321_ec2_common_tags, {
    pci_item_id      = "ec2_3"
    pci_requirements = "2.4"
  })
}

control "pci_v321_ec2_4" {
  title         = "4 Unused EC2 EIPs should be removed"
  description   = "This control checks whether Elastic IP addresses that are allocated to a VPC are attached to Amazon EC2 instances or in-use elastic network interfaces (ENIs). A failed finding indicates you may have unused Amazon EC2 EIPs. This will help you maintain an accurate asset inventory of EIPs in your cardholder data environment (CDE)."
  severity      = "low"
  sql           = query.vpc_eip_associated.sql
  documentation = file("./pci_v321/docs/pci_v321_ec2_4.md")

  tags = merge(local.pci_v321_ec2_common_tags, {
    pci_item_id      = "ec2_4"
    pci_requirements = "2.4"
  })
}

control "pci_v321_ec2_5" {
  title         = "5 Security groups should not allow ingress from 0.0.0.0/0 to port 22"
  description   = "This control checks whether security groups in use disallow unrestricted incoming SSH traffic. It does not evaluate outbound traffic. Note that security groups are stateful. If you send a request from your instance, the response traffic for that request is allowed to flow in regardless of inbound security group rules. Responses to allowed inbound traffic are allowed to flow out regardless of outbound rules."
  severity      = "high"
  sql           = query.vpc_security_group_remote_administration.sql
  documentation = file("./pci_v321/docs/pci_v321_ec2_5.md")

  tags = merge(local.pci_v321_ec2_common_tags, {
    pci_item_id      = "ec2_5"
    pci_requirements = "1.2.1,1.3.1,2.2.2"
  })
}

control "pci_v321_ec2_6" {
  title         = "6 VPC flow logging should be enabled in all VPCs"
  description   = "This control checks whether VPC flow logs are found and enabled for VPCs. The traffic type is set to REJECT. With VPC Flow Logs, you can capture information about the IP address traffic to and from network interfaces in your VPC. After you create a flow log, you can use CloudWatch Logs to view and retrieve the log data. Security Hub recommends that you enable flow logging for packet rejects for VPCs. Flow logs provide visibility into network traffic that traverses the VPC. They can detect anomalous traffic and provide insight into security workflows. By default, the record includes values for the different components of the IP address flow, including the source, destination, and protocol."
  severity      = "medium"
  sql           = query.vpc_flow_logs_enabled.sql
  documentation = file("./pci_v321/docs/pci_v321_ec2_6.md")

  tags = merge(local.pci_v321_ec2_common_tags, {
    pci_item_id      = "ec2_6"
    pci_requirements = "10.3.3,10.3.4,10.3.5,10.3.6"
  })
}
