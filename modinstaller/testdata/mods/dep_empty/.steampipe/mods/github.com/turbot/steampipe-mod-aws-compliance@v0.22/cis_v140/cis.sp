locals {
  cis_v140_common_tags = {
    cis         = "true"
    cis_version = "v1.4.0"
    plugin      = "aws"
  }
}

benchmark "cis_v140" {
  title         = "CIS v1.4.0"
  description   = "The CIS Amazon Web Services Foundations Benchmark provides prescriptive guidance for configuring security options for a subset of Amazon Web Services with an emphasis on foundational, testable, and architecture agnostic settings."
  documentation = file("./cis_v140/docs/cis_overview.md")
  children = [
    benchmark.cis_v140_1,
    benchmark.cis_v140_2,
    benchmark.cis_v140_3,
    benchmark.cis_v140_4,
    benchmark.cis_v140_5
  ]
  tags = local.cis_v140_common_tags
}
