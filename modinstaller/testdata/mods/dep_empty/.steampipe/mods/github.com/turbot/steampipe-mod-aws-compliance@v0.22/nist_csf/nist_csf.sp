locals {
  nist_csf_common_tags = {
    nist_csf = "true"
    plugin   = "aws"
  }
}

benchmark "nist_csf" {
  title         = "NIST Cybersecurity Framework (CSF) v1.1"
  description   = "NIST Cybersecurity Framework is a set of best practices, standards, and recommendations that help an organization improve its cybersecurity measures."
  documentation = file("./nist_csf/docs/nist_csf_overview.md")

  children = [
    benchmark.nist_csf_de,
    benchmark.nist_csf_id,
    benchmark.nist_csf_pr,
    benchmark.nist_csf_rc,
    benchmark.nist_csf_rs
  ]

  tags = local.nist_csf_common_tags
}
