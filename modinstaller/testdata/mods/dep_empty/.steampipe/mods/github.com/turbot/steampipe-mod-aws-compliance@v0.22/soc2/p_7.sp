locals {
  soc_2_p_7_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "p7"
  })
}

benchmark "soc_2_p_7" {
  title       = "P7.0 - Privacy Criteria Related to Quality"
  description = "This category refers to privacy criteria related to quality."

  children = [
    benchmark.soc_2_p_7_1
  ]

  tags = local.soc_2_p_7_common_tags
}

benchmark "soc_2_p_7_1" {
  title         = "P7.1 The entity collects and maintains accurate, up-to-date, complete, and relevant personal information to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_7_1.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_7_common_tags, {
    soc_2_item_id = "7.1"
    soc_2_type    = "manual"
  })
}