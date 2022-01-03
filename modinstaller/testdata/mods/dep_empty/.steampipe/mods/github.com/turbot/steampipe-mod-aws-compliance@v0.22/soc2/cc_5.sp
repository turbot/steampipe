locals {
  soc_2_cc_5_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "cc5"
  })
}

benchmark "soc_2_cc_5" {
  title       = "CC5.0 - Control Activities"
  description = "The criteria relevant to how the entity (i) selects and develops control activities, (ii) selects and develops general controls over technology, and (iii) deploys through policies and procedures."

  children = [
    benchmark.soc_2_cc_5_1,
    benchmark.soc_2_cc_5_2,
    benchmark.soc_2_cc_5_3
  ]

  tags = local.soc_2_cc_5_common_tags
}

benchmark "soc_2_cc_5_1" {
  title         = "CC5.1 COSO Principle 10: The entity selects and develops control activities that contribute to the mitigation of risks to the achievement of objectives to acceptable levels"
  documentation = file("./soc2/docs/cc_5_1.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_5_common_tags, {
    soc_2_item_id = "5.1"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_cc_5_2" {
  title         = "CC5.2 COSO Principle 11: The entity also selects and develops general control activities over technology to support the achievement of objectives"
  documentation = file("./soc2/docs/cc_5_2.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_5_common_tags, {
    soc_2_item_id = "5.2"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_cc_5_3" {
  title         = "CCC5.3 COSO Principle 12: The entity deploys control activities through policies that establish what is expected and in procedures that put policies into action"
  documentation = file("./soc2/docs/cc_5_3.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_5_common_tags, {
    soc_2_item_id = "5.3"
    soc_2_type    = "manual"
  })
}