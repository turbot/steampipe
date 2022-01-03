benchmark "hipaa_164_308_a_5_ii_b" {
  title       = "164.308(a)(5)(ii)(B) Protection from malicious software"
  description = "Procedures for guarding against, detecting, and reporting malicious software."
  children = [
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant,
    control.ssm_managed_instance_compliance_patch_compliant
  ]

  tags = merge(local.hipaa_164_308_common_tags, {
    hipaa_item_id = "164_308_a_5_ii_b"
  })
}