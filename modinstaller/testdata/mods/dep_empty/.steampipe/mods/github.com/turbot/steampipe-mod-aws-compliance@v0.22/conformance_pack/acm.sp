locals {
  conformance_pack_acm_common_tags = {
    service = "acm"
  }
}

control "acm_certificate_expires_30_days" {
  title       = "ACM certificates should be set to expire within 30 days"
  description = "Ensure network integrity is protected by ensuring X509 certificates are issued by AWS ACM."
  sql         = query.acm_certificate_expires_30_days.sql

  tags = merge(local.conformance_pack_acm_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}