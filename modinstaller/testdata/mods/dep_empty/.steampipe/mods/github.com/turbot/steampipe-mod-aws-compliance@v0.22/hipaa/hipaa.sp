locals {
  hipaa_common_tags = {
    hipaa  = "true"
    plugin = "aws"
  }
}

benchmark "hipaa" {
  title       = "HIPAA"
  description = "The AWS Health Insurance Portability and Accountability (HIPAA) is a set of controls to use the secure AWS environment to process, maintain, and store protected health information."
  children = [
    benchmark.hipaa_164_308,
    benchmark.hipaa_164_312
  ]

  tags = local.hipaa_common_tags
}
