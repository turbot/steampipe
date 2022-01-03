locals {
  foundational_security_ec2_common_tags = merge(local.foundational_security_common_tags, {
    service = "ec2"
  })
}

benchmark "foundational_security_ec2" {
  title         = "EC2"
  documentation = file("./foundational_security/docs/foundational_security_ec2.md")
  children = [
    control.foundational_security_ec2_1,
    control.foundational_security_ec2_2,
    control.foundational_security_ec2_3,
    control.foundational_security_ec2_4,
    control.foundational_security_ec2_6,
    control.foundational_security_ec2_7,
    control.foundational_security_ec2_8,
    control.foundational_security_ec2_9,
    control.foundational_security_ec2_10,
    control.foundational_security_ec2_15,
    control.foundational_security_ec2_16,
    control.foundational_security_ec2_17,
    control.foundational_security_ec2_18,
    control.foundational_security_ec2_19
  ]
  tags          = local.foundational_security_ec2_common_tags
}

control "foundational_security_ec2_1" {
  title         = "1 Amazon EBS snapshots should not be public, determined by the ability to be restorable by anyone"
  description   = "Amazon EBS snapshots should not be public, determined by the ability to be restorable by anyone. EBS snapshots are used to back up the data on your EBS volumes to Amazon S3 at a specific point in time. You can use the snapshots to restore previous states of EBS volumes. It is rarely acceptable to share a snapshot with the public. Typically the decision to share a snapshot publicly was made in error or without a complete understanding of the implications. This check helps ensure that all such sharing was fully planned and intentional."
  severity      = "critical"
  sql           = query.ebs_snapshot_not_publicly_restorable.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_1.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_1"
    foundational_security_category = "secure_network_configuration"
  })
}

control "foundational_security_ec2_2" {
  title         = "2 The VPC default security group should not allow inbound and outbound traffic"
  description   = "This control checks that the default security group of a VPC does not allow inbound or outbound traffic. The rules for the default security group allow all outbound and inbound traffic from network interfaces (and their associated instances) that are assigned to the same security group."
  severity      = "high"
  sql           = query.vpc_default_security_group_restricts_all_traffic.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_2.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_2"
    foundational_security_category = "secure_network_configuration"
  })
}

control "foundational_security_ec2_3" {
  title         = "3 Attached EBS volumes should be encrypted at rest"
  description   = "This control checks whether the EBS volumes that are in an attached state are encrypted. To pass this check, EBS volumes must be in use and encrypted. If the EBS volume is not attached, then it is not subject to this check."
  severity      = "medium"
  sql           = query.ebs_attached_volume_encryption_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_3.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_3"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}

control "foundational_security_ec2_4" {
  title         = "4 Stopped EC2 instances should be removed after a specified time period"
  description   = "This control checks whether any EC2 instances have been stopped for more than the allowed number of days. An EC2 instance fails this check if it is stopped for longer than the maximum allowed time period, which by default is 30 days."
  severity      = "medium"
  sql           = query.ec2_stopped_instance_30_days.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_4.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_4"
    foundational_security_category = "inventory"
  })
}

control "foundational_security_ec2_6" {
  title         = "6 VPC flow logging should be enabled in all VPCs"
  description   = "This control checks whether Amazon VPC Flow Logs are found and enabled for VPCs. The traffic type is set to Reject. With the VPC Flow Logs feature, you can capture information about the IP address traffic going to and from network interfaces in your VPC. After you create a flow log, you can view and retrieve its data in CloudWatch Logs. To reduce cost, you can also send your flow logs to Amazon S3."
  severity      = "medium"
  sql           = query.vpc_flow_logs_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_6.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_6"
    foundational_security_category = "logging"
  })
}

control "foundational_security_ec2_7" {
  title         = "7 EBS default encryption should be enabled"
  description   = "This control checks whether account-level encryption is enabled by default for Amazon Elastic Block Store(Amazon EBS). The control fails if the account level encryption is not enabled."
  severity      = "medium"
  sql           = query.ec2_ebs_default_encryption_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_7.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_7"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}

control "foundational_security_ec2_8" {
  title         = "8 EC2 instances should use IMDSv2"
  description   = "This control checks whether your EC2 instance metadata version is configured with Instance Metadata Service Version 2 (IMDSv2). The control passes if HttpTokens is set to required for IMDSv2. The control fails if HttpTokens is set to optional."
  severity      = "high"
  sql           = query.ec2_instance_uses_imdsv2.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_8.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_8"
    foundational_security_category = "network_security"
  })
}

control "foundational_security_ec2_9" {
  title         = "9 EC2 instances should not have a public IP address"
  description   = "This control checks whether EC2 instances have a public IP address. The control fails if the publicIp field is present in the EC2 instance configuration item. This control applies to IPv4 addresses only."
  severity      = "high"
  sql           = query.ec2_instance_not_publicly_accessible.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_9.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_9"
    foundational_security_category = "public_ip_addresses"
  })
}

control "foundational_security_ec2_10" {
  title         = "10 Amazon EC2 should be configured to use VPC endpoints"
  description   = "This control checks whether a service endpoint for Amazon EC2 is created for each VPC. The control fails if a VPC does not have a VPC endpoint created for the Amazon EC2 service."
  severity      = "medium"
  sql           = query.vpc_configured_to_use_vpc_endpoints.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_10.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_10"
    foundational_security_category = "api_private_access"
  })
}

control "foundational_security_ec2_15" {
  title         = "15 EC2 subnets should not automatically assign public IP addresses"
  description   = "This control checks whether the assignment of public IPs in Amazon Virtual Private Cloud (Amazon VPC) subnets have MapPublicIpOnLaunch set to FALSE. The control passes if the flag is set to FALSE."
  severity      = "medium"
  sql           = query.vpc_subnet_auto_assign_public_ip_disabled.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_15.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_15"
    foundational_security_category = "network_security"
  })
}

control "foundational_security_ec2_16" {
  title         = "16 Unused network access control lists should be removed"
  description   = "This control checks whether there are any unused network access control lists (ACLs). The control checks the item configuration of the resource AWS::EC2::NetworkAcl and determines the relationships of the network ACL."
  severity      = "low"
  sql           = query.vpc_network_acl_unused.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_16.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_16"
    foundational_security_category = "network_security"
  })
}

control "foundational_security_ec2_17" {
  title         = "17 EC2 instances should not use multiple ENIs"
  description   = "This control checks whether an EC2 instance uses multiple Elastic Network Interfaces (ENIs) or Elastic Fabric Adapters (EFAs). This control passes if a single network adapter is used. The control includes an optional parameter list to identify the allowed ENIs."
  severity      = "low"
  sql           = query.ec2_instance_not_use_multiple_enis.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_17.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_17"
    foundational_security_category = "network_security"
  })
}

control "foundational_security_ec2_18" {
  title         = "18 Security groups should only allow unrestricted incoming traffic for authorized ports"
  description   = "This control checks whether the security groups that are in use allow unrestricted incoming traffic. Optionally the rule checks whether the port numbers are listed in the authorizedTcpPorts parameter. The default values for authorizedTcpPorts are 80 and 443."
  severity      = "high"
  sql           = query.vpc_security_group_allows_ingress_authorized_ports.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_18.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_18"
    foundational_security_category = "security_group_configuration"
  })
}

control "foundational_security_ec2_19" {
  title         = "19 Security groups should not allow unrestricted access to ports with high risk"
  description   = "This control checks whether unrestricted incoming traffic for the security groups is accessible to the specified ports that have the highest risk. This control passes when none of the rules in a security group allow ingress traffic from 0.0.0.0/0 for those ports."
  severity      = "medium"
  sql           = query.vpc_security_group_restricted_common_ports.sql
  documentation = file("./foundational_security/docs/foundational_security_ec2_19.md")

  tags = merge(local.foundational_security_ec2_common_tags, {
    foundational_security_item_id  = "ec2_19"
    foundational_security_category = "security_group_configuration"
  })
}