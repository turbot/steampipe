locals {
  nist_800_53_rev_4_common_tags = {
    nist_800_53_rev_4 = "true"
    plugin            = "aws"
  }
}

benchmark "nist_800_53_rev_4" {
  title       = "NIST 800-53 Revision 4"
  description = "NIST 800-53 is a regulatory standard that defines the minimum baseline of security controls for all U.S. federal information systems except those related to national security."
  documentation = file("./nist_800_53_rev_4/docs/nist_800_53_rev_4_overview.md")

  children = [
    benchmark.nist_800_53_rev_4_ac,
    benchmark.nist_800_53_rev_4_au,
    benchmark.nist_800_53_rev_4_ca,
    benchmark.nist_800_53_rev_4_cm,
    benchmark.nist_800_53_rev_4_cp,
    benchmark.nist_800_53_rev_4_ia,
    benchmark.nist_800_53_rev_4_ir,
    benchmark.nist_800_53_rev_4_ra,
    benchmark.nist_800_53_rev_4_sa,
    benchmark.nist_800_53_rev_4_sc,
    benchmark.nist_800_53_rev_4_si
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
