locals {
  soc_2_cc_2_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "cc2"
  })
}

benchmark "soc_2_cc_2" {
  title       = "CC2.0 - Common Criteria Related to Communication and Information"
  description = "The criteria relevant to how the entity (i) uses relevant information, (ii) communicates internally, and (iii) communicates externally."

  children = [
    benchmark.soc_2_cc_2_1,
    benchmark.soc_2_cc_2_2,
    benchmark.soc_2_cc_2_3,
  ]

  tags = local.soc_2_cc_2_common_tags
}

benchmark "soc_2_cc_2_1" {
  title       = "CC2.1 COSO Principle 13: The entity obtains or generates and uses relevant, quality information to support the functioning of internal control"
  documentation = file("./soc2/docs/cc_2_1.md")

  children = [
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_security_trail_enabled,
    control.config_enabled_all_regions,
  ]

  tags = merge(local.soc_2_cc_2_common_tags, {
    soc_2_item_id = "2.1"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_2_2" {
  title         = "CC2.2 COSO Principle 14: The entity internally communicates information, including objectives and responsibilities for internal control, necessary to support the functioning of internal control"
  documentation = file("./soc2/docs/cc_2_2.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_2_common_tags, {
    soc_2_item_id = "2.2"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_cc_2_3" {
  title         = "CC2.3 COSO Principle 15: The entity communicates with external parties regarding matters affecting the functioning of internal control"
  documentation = file("./soc2/docs/cc_2_3.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_2_common_tags, {
    soc_2_item_id = "2.3"
    soc_2_type    = "manual"
  })
}