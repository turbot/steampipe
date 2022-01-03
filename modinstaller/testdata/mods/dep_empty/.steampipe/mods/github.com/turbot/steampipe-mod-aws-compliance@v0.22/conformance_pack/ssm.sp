locals {
  conformance_pack_ssm_common_tags = {
    service = "ssm"
  }
}

control "ec2_instance_ssm_managed" {
  title       = "EC2 instances should be managed by AWS Systems Manager"
  description = "An inventory of the software platforms and applications within the organization is possible by managing Amazon Elastic Compute Cloud (Amazon EC2) instances with AWS Systems Manager."
  sql         = query.ec2_instance_ssm_managed.sql

  tags = merge(local.conformance_pack_ssm_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "ssm_managed_instance_compliance_association_compliant" {
  title       = "SSM managed instance associations should be compliant"
  description = "Use AWS Systems Manager Associations to help with inventory of software platforms and applications within an organization."
  sql         = query.ssm_managed_instance_compliance_association_compliant.sql

  tags = merge(local.conformance_pack_ssm_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "ssm_managed_instance_compliance_patch_compliant" {
  title       = "SSM managed instance patching should be compliant"
  description = "Enable this rule to help with identification and documentation of Amazon Elastic Compute Cloud (Amazon EC2) vulnerabilities."
  sql         = query.ssm_managed_instance_compliance_patch_compliant.sql

  tags = merge(local.conformance_pack_ssm_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}
