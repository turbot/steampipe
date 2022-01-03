locals {
  soc_2_p_4_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "p4"
  })
}

benchmark "soc_2_p_4" {
  title       = "P4.0 - Privacy Criteria Related to Use, Retention, and Disposal"
  description = "This category refers to privacy criteria related to use, retention, and disposal."

  children = [
    benchmark.soc_2_p_4_1,
    benchmark.soc_2_p_4_2,
    benchmark.soc_2_p_4_3
  ]

  tags = local.soc_2_p_4_common_tags
}

benchmark "soc_2_p_4_1" {
  title         = "P4.1 The entity limits the use of personal information to the purposes identified in the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_4_1.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_4_common_tags, {
    soc_2_item_id = "4.1"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_4_2" {
  title         = "P4.2 The entity retains personal information consistent with the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_4_2.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_4_common_tags, {
    soc_2_item_id = "4.2"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_4_3" {
  title         = "P4.3 The entity securely disposes of personal information to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_4_3.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_4_common_tags, {
    soc_2_item_id = "4.3"
    soc_2_type    = "manual"
  })
}
