locals {
  soc_2_p_5_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "p5"
  })
}

benchmark "soc_2_p_5" {
  title       = "P5.0 - Privacy Criteria Related to Access"
  description = "This category refers to privacy criteria related to access."

  children = [
    benchmark.soc_2_p_5_1,
    benchmark.soc_2_p_5_2
  ]

  tags = local.soc_2_p_5_common_tags
}

benchmark "soc_2_p_5_1" {
  title         = "P5.1 The entity grants identified and authenticated data subjects the ability to access their stored personal information for review and, upon request, provides physical or electronic copies of that information to data subjects to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_5_1.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_5_common_tags, {
    soc_2_item_id = "5.1"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_5_2" {
  title         = "P5.2 The entity corrects, amends, or appends personal information based on information provided by data subjects and communicates such information to third parties, as committed or required, to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_5_2.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_5_common_tags, {
    soc_2_item_id = "5.2"
    soc_2_type    = "manual"
  })
}
