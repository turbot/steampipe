locals {
  soc_2_cc_6_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "cc6"
  })
}

benchmark "soc_2_cc_6" {
  title       = "CC6.0 - Logical and Physical Access"
  description = "The criteria relevant to how an entity (i) restricts logical and physical access, (ii) provides and removes that access, and (iii) prevents unauthorized access."

  children = [
    benchmark.soc_2_cc_6_1,
    benchmark.soc_2_cc_6_2,
    benchmark.soc_2_cc_6_3,
    benchmark.soc_2_cc_6_4,
    benchmark.soc_2_cc_6_5,
    benchmark.soc_2_cc_6_6,
    benchmark.soc_2_cc_6_7,
    benchmark.soc_2_cc_6_8
  ]

  tags = local.soc_2_cc_6_common_tags
}

benchmark "soc_2_cc_6_1" {
  title         = "CC6.1 The entity implements logical access security software, infrastructure, and architectures over protected information assets to protect them from security events to meet the entity's objectives"
  documentation = file("./soc2/docs/cc_6_1.md")

  children = [
   control.s3_bucket_restrict_public_read_access
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.1"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_6_2" {
  title         = "CC6.2 Prior to issuing system credentials and granting system access, the entity registers and authorizes new internal and external users whose access is administered by the entity"
  documentation = file("./soc2/docs/cc_6_2.md")

  children = [
    control.rds_db_instance_prohibit_public_access
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.2"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_6_3" {
  title         = "CC6.3 The entity authorizes, modifies, or removes access to data, software, functions, and other protected information assets based on roles, responsibilities, or the system design and changes, giving consideration to the concepts of least privilege and segregation of duties, to meet the entity’s objectives"
  documentation = file("./soc2/docs/cc_6_3.md")

  children = [
    control.iam_policy_no_star_star
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.3"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_6_4" {
  title         = "CC6.4 The entity restricts physical access to facilities and protected information assets to authorized personnel to meet the entity’s objectives"
  documentation = file("./soc2/docs/cc_6_4.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.4"
    soc_2_type    = "manual"
  })
}

benchmark "soc_2_cc_6_5" {
  title         = "CC6.5 The entity discontinues logical and physical protections over physical assets only after the ability to read or recover data and software from those assets has been diminished and is no longer required to meet the entity’s objectives"
  documentation = file("./soc2/docs/cc_6_5.md")

  children = [
    control.manual_control
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.5"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_6_6" {
  title         = "CC6.6 The entity implements logical access security measures to protect against threats from sources outside its system boundaries"
  documentation = file("./soc2/docs/cc_6_6.md")

  children = [
    control.ec2_instance_not_publicly_accessible
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.6"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_6_7" {
  title         = "CC6.7 The entity restricts the transmission, movement, and removal of information to authorized internal and external users and processes, and protects it during transmission, movement, or removal to meet the entity’s objectives"
  documentation = file("./soc2/docs/cc_6_7.md")

  children = [
    control.acm_certificate_expires_30_days
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.7"
    soc_2_type    = "automated"
  })
}

benchmark "soc_2_cc_6_8" {
  title         = "CC6.8 The entity implements controls to prevent or detect and act upon the introduction of unauthorized or malicious software to meet the entity’s objectives"
  documentation = file("./soc2/docs/cc_6_8.md")

  children = [
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = merge(local.soc_2_cc_6_common_tags, {
    soc_2_item_id = "6.8"
    soc_2_type    = "automated"
  })
}
