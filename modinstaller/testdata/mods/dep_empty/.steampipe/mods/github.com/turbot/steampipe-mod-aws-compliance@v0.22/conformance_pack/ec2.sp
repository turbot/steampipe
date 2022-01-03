locals {
  conformance_pack_ec2_common_tags = {
    service = "ec2"
  }
}

control "ec2_ebs_default_encryption_enabled" {
  title       = "EBS default encryption should be enabled"
  description = "To help protect data at rest, ensure that encryption is enabled for your Amazon Elastic Block Store (Amazon EBS) volumes."
  sql         = query.ec2_ebs_default_encryption_enabled.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    hipaa = "true"
  })
}

control "ec2_instance_detailed_monitoring_enabled" {
  title       = "EC2 instance detailed monitoring should be enabled"
  description = "Enable this rule to help improve Amazon Elastic Compute Cloud (Amazon EC2) instance monitoring on the Amazon EC2 console, which displays monitoring graphs with a 1-minute period for the instance."
  sql         = query.ec2_instance_detailed_monitoring_enabled.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
    soc_2             = "true"
  })
}

control "ec2_instance_in_vpc" {
  title       = "EC2 instances should be in a VPC"
  description = "Deploy Amazon Elastic Compute Cloud (Amazon EC2) instances within an Amazon Virtual Private Cloud (Amazon VPC) to enable secure communication between an instance and other services within the amazon VPC, without requiring an internet gateway, NAT device, or VPN connection."
  sql         = query.ec2_instance_in_vpc.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "ec2_instance_not_publicly_accessible" {
  title       = "EC2 instances should not have a public IP address"
  description = "Manage access to the AWS Cloud by ensuring Amazon Elastic Compute Cloud (Amazon EC2) instances cannot be publicly accessed."
  sql         = query.ec2_instance_not_publicly_accessible.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "ec2_stopped_instance_30_days" {
  title       = "EC2 stopped instances should be removed in 30 days"
  description = "Enable this rule to help with the baseline configuration of Amazon Elastic Compute Cloud (Amazon EC2) instances by checking whether Amazon EC2 instances have been stopped for more than the allowed number of days, according to your organization's standards."
  sql         = query.ec2_stopped_instance_30_days.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
  })
}

control "ec2_instance_ebs_optimized" {
  title       = "EC2 instance should have EBS optimization enabled"
  description = "An optimized instance in Amazon Elastic Block Store (Amazon EBS) provides additional, dedicated capacity for Amazon EBS I/O operations."
  sql         = query.ec2_instance_ebs_optimized.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    audit_manager_control_tower = "true"
    hipaa                       = "true"
    nist_csf                    = "true"
    soc_2                       = "true"
  })
}

control "ec2_instance_uses_imdsv2" {
  title       = "EC2 instances should use IMDSv2"
  description = "Ensure the Instance Metadata Service Version 2 (IMDSv2) method is enabled to help protect access and control of Amazon Elastic Compute Cloud (Amazon EC2) instance metadata."
  sql         = query.ec2_instance_uses_imdsv2.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
  })
}

control "ec2_instance_protected_by_backup_plan" {
  title       = "EC2 instances should be protected by backup plan"
  description = "Ensure if Amazon Elastic Compute Cloud (Amazon EC2) instances are protected by a backup plan. The rule is non complaint if the Amazon EC2 instance is not covered by a backup plan."
  sql         = query.ec2_instance_protected_by_backup_plan.sql

  tags = merge(local.conformance_pack_ec2_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}
