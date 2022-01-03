locals {
  soc_2_p_1_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "p1"
  })
}

benchmark "soc_2_p_1" {
  title       = "P1.0 - Privacy Criteria Related to Notice and Communication of Objectives Related to Privacy"
  description = "This category refers to privacy criteria related to notice and communication of objectives related to privacy."

  children = [
    benchmark.soc_2_p_1_1
  ]

  tags = local.soc_2_p_1_common_tags
}

benchmark "soc_2_p_1_1" {
  title         = "P1.1 The entity provides notice to data subjects about its privacy practices to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_1_1.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_1_common_tags, {
    soc_2_item_id = "1.1"
    soc_2_type    = "manual"
  })
}
