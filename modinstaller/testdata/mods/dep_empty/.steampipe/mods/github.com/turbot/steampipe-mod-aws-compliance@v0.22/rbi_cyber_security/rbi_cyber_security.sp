locals {
  rbi_cyber_security_common_tags = {
    rbi_cyber_security = "true"
    plugin             = "aws"
  }
}

benchmark "rbi_cyber_security" {
  title       = "RBI Cyber Security Framework"
  description = "The Reserve Bank had prescribed a set of baseline cyber security controls for primary (Urban) cooperative banks (UCBs) in October 2018. On further examination, it has been decided to prescribe a comprehensive cyber security framework for the UCBs, as a graded approach, based on their digital depth and interconnectedness with the payment systems landscape, digital products offered by them and assessment of cyber security risk. The framework would mandate implementation of progressively stronger security measures based on the nature, variety and scale of digital product offerings of banks."
  children = [
    benchmark.rbi_cyber_security_annex_i_1_1,
    benchmark.rbi_cyber_security_annex_i_1_3,
    benchmark.rbi_cyber_security_annex_i_5_1,
    benchmark.rbi_cyber_security_annex_i_6,
    benchmark.rbi_cyber_security_annex_i_7_1,
    benchmark.rbi_cyber_security_annex_i_7_2,
    benchmark.rbi_cyber_security_annex_i_7_3,
    benchmark.rbi_cyber_security_annex_i_7_4,
    benchmark.rbi_cyber_security_annex_i_12
  ]
  tags = local.rbi_cyber_security_common_tags
}
