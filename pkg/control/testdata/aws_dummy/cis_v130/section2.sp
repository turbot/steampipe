locals {
  cis_v130_2_common_tags = merge(local.cis_v130_common_tags, {
    cis_section_id = "2"
  })
}

locals {
  cis_v130_2_1_common_tags = merge(local.cis_v130_2_common_tags, {
    cis_section_id = "2.1"
  })
  cis_v130_2_2_common_tags = merge(local.cis_v130_2_common_tags, {
    cis_section_id = "2.2"
  })
}

benchmark "cis_v130_2" {
  title         = "2 Storage"
  documentation = file("./cis_v130/docs/cis_v130_2.md")
  children = [
    benchmark.cis_v130_2_1,
    benchmark.cis_v130_2_2
  ]

  tags          = local.cis_v130_2_common_tags
}

benchmark "cis_v130_2_1" {
  title         = "2.1 Simple Storage Service (S3)"
  documentation = file("./cis_v130/docs/cis_v130_2_1.md")
  children = [
    control.cis_v130_2_1_1,
    control.cis_v130_2_1_2
  ]
  tags          = local.cis_v130_2_1_common_tags
}

benchmark "cis_v130_2_2" {
  title         = "2.2 Elastic Compute Cloud (EC2)"
  documentation = file("./cis_v130/docs/cis_v130_2_2.md")
  children = [
    control.cis_v130_2_2_1
  ]
  tags          = local.cis_v130_2_2_common_tags
}

control "cis_v130_2_1_1" {
  title         = "2.1.1 Ensure all S3 buckets employ encryption-at-rest"
  description   = "Amazon S3 provides a variety of no, or low, cost encryption options to protect data at rest."
  documentation = file("./cis_v130/docs/cis_v130_2_1_1.md")
  sql           = query.ok.sql

  tags = merge(local.cis_v130_2_1_common_tags, {
    cis_item_id  = "2.1.1"
    cis_type     = "manual"
    cis_levels   = "1,2"
    cis_controls = "14.8"
  })
}

control "cis_v130_2_1_2" {
  title         = "2.1.2 Ensure S3 Bucket Policy allows HTTPS requests"
  description   = "At the Amazon S3 bucket level, you can configure permissions through a bucket policy making the objects accessible only through HTTPS."
  documentation = file("./cis_v130/docs/cis_v130_2_1_2.md")
  sql           = query.info.sql

  tags = merge(local.cis_v130_2_1_common_tags, {
    cis_item_id  = "2.1.2"
    cis_type     = "manual"
    cis_levels   = "1,2"
    cis_controls = "14.8"
  })
}

control "cis_v130_2_2_1" {
  title       = "2.2.1 Ensure EBS volume encryption is enabled"
  description = "Elastic Compute Cloud (EC2) supports encryption at rest when using the Elastic Block Store (EBS) service. While disabled by default, forcing encryption at EBS volume creation is supported."
  #documentation = file("./cis_v130/docs/cis_v130_2_2_1.md")
  sql = query.ok.sql

  tags = merge(local.cis_v130_2_2_common_tags, {
    cis_item_id  = "2.2.1"
    cis_type     = "manual"
    cis_levels   = "1,2"
    cis_controls = "14.8"
  })
}
