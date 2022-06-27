locals {
  cis_v130_common_tags = {
    benchmark            = "cis"
    cis_controls_version = "v7.1"
    cis_version          = "v1.3.0"
    plugin               = "aws"
  }
}

benchmark "cis_v130" {
  title         = "CIS v1.3.0"
  description   = "The CIS Amazon Web Services Foundations Benchmark provides prescriptive guidance for configuring security options for a subset of Amazon Web Services with an emphasis on foundational, testable, and architecture agnostic settings."
  documentation = file("./cis_v130/docs/cis-overview.md")
  children = [
    benchmark.cis_v130_1,
    benchmark.cis_v130_2,
    benchmark.cis_v130_3,
    benchmark.cis_v130_4,
    benchmark.cis_v130_5
  ]
  tags = local.cis_v130_common_tags
}
