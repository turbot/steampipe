locals {
  soc_2_cc_3_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "cc3"
  })
}

benchmark "soc_2_cc_3" {
  title       = "CC3.0 - Common Criteria Related to Risk Assessment"
  description = "The criteria relevant to how the entity (i) specifies suitable objectives, (ii) identifies and analyzes risk, and (iii) assess fraud risk."

  children = [
    benchmark.soc_2_cc_3_1,
    benchmark.soc_2_cc_3_2,
    benchmark.soc_2_cc_3_3,
    benchmark.soc_2_cc_3_4
  ]

  tags = local.soc_2_cc_3_common_tags
}

benchmark "soc_2_cc_3_1" {
  title         = "CC3.1 COSO Principle 6: The entity specifies objectives with sufficient clarity to enable the identification and assessment of risks relating to objectives"
  documentation = file("./soc2/docs/cc_3_1.md")

  children = [
    control.guardduty_enabled,
    control.securityhub_enabled,
    control.config_enabled_all_regions
  ]

  tags = merge(local.soc_2_cc_3_common_tags, {
    soc_2_item_id = "3.1"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_3_2" {
  title         = "CC3.2 COSO Principle 7: The entity identifies risks to the achievement of its objectives across the entity and analyzes risks as a basis for determining how the risks should be managed"
  documentation = file("./soc2/docs/cc_3_2.md")

  children = [
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant,
    control.guardduty_finding_archived,
    control.guardduty_enabled,
    control.ssm_managed_instance_compliance_patch_compliant
  ]

  tags = merge(local.soc_2_cc_3_common_tags, {
    soc_2_item_id = "3.2"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_3_3" {
  title         = "CC3.3 COSO Principle 8: The entity considers the potential for fraud in assessing risks to the achievement of objectives"
  documentation = file("./soc2/docs/cc_3_3.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_4_common_tags, {
    soc_2_item_id = "3.3"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_cc_3_4" {
  title         = "CC3.4 COSO Principle 9: The entity identifies and assesses changes that could significantly impact the system of internal control"
  documentation = file("./soc2/docs/cc_3_4.md")

  children = [
    control.config_enabled_all_regions
  ]

  tags = merge(local.soc_2_cc_3_common_tags, {
    soc_2_item_id = "3.4"
    soc_2_type    = "automated"
  })
}
