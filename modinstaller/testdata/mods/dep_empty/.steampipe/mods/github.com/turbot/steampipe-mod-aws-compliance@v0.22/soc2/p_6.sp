locals {
  soc_2_p_6_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "p6"
  })
}

benchmark "soc_2_p_6" {
  title       = "P6.0 - Privacy Criteria Related to Disclosure and Notification"
  description = "This category refers to privacy criteria related to disclosure and notification."

  children = [
    benchmark.soc_2_p_6_1,
    benchmark.soc_2_p_6_2,
    benchmark.soc_2_p_6_3,
    benchmark.soc_2_p_6_4,
    benchmark.soc_2_p_6_5,
    benchmark.soc_2_p_6_6,
    benchmark.soc_2_p_6_7
  ]

  tags = local.soc_2_p_6_common_tags
}

benchmark "soc_2_p_6_1" {
  title         = "P6.1 The entity discloses personal information to third parties with the explicit consent of data subjects, and such consent is obtained prior to disclosure to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_6_1.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.1"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_6_2" {
  title         = "P6.2 The entity creates and retains a complete, accurate, and timely record of authorized disclosures of personal information to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_6_2.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.2"
    soc_2_type    = "manual"
  })
}


benchmark "soc_2_p_6_3" {
  title         = "P6.3 The entity creates and retains a complete, accurate, and timely record of detected or reported unauthorized disclosures (including breaches) of personal information to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_6_3.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.3"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_6_4" {
  title         = "P6.4 The entity obtains privacy commitments from vendors and other third parties who have access to personal information to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_6_4.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.4"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_6_5" {
  title         = "P6.5 The entity obtains commitments from vendors and other third parties with access to personal information to notify the entity in the event of actual or suspected unauthorized disclosures of personal information"
  documentation = file("./soc2/docs/p_6_5.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.5"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_6_6" {
  title         = "P6.6 The entity provides notification of breaches and incidents to affected data subjects, regulators, and others to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_6_6.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.6"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_p_6_7" {
  title         = "P6.7 The entity provides data subjects with an accounting of the personal information held and disclosure of the data subjects’ personal information, upon the data subjects’ request, to meet the entity’s objectives related to privacy"
  documentation = file("./soc2/docs/p_6_7.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_p_6_common_tags, {
    soc_2_item_id = "6.7"
    soc_2_type    = "manual"
  })
}