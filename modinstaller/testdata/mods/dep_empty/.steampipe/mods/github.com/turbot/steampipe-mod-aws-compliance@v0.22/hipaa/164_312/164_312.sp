locals {
  hipaa_164_312_common_tags = merge(local.hipaa_common_tags, {
    hipaa_section = "164_312"
  })
}

benchmark "hipaa_164_312" {
  title       = "164.312 Technical Safeguards"
  description = "The Security Rule defines technical safeguards in § 164.304 as `the technology and the policy and procedures for its use that protect electronic protected health information and control access to it.`"

  children = [
    benchmark.hipaa_164_312_a_1,
    benchmark.hipaa_164_312_a_2_i,
    benchmark.hipaa_164_312_a_2_ii,
    benchmark.hipaa_164_312_a_2_iv,
    benchmark.hipaa_164_312_b,
    benchmark.hipaa_164_312_c_1,
    benchmark.hipaa_164_312_c_2,
    benchmark.hipaa_164_312_d,
    benchmark.hipaa_164_312_e_1,
    benchmark.hipaa_164_312_e_2_i,
    benchmark.hipaa_164_312_e_2_ii
  ]

  tags = local.hipaa_164_312_common_tags
}
