locals {
  conformance_pack_wafv2_common_tags = {
    service = "wafv2"
  }
}

control "wafv2_web_acl_logging_enabled" {
  title       = "Logging should be enabled on AWS WAFv2 regional and global web access control list (ACLs)"
  description = "To help with logging and monitoring within your environment, enable AWS WAF (V2) logging on regional and global web ACLs."
  sql         = query.wafv2_web_acl_logging_enabled.sql

  tags = merge(local.conformance_pack_wafv2_common_tags, {
    hipaa              = "true"
    gdpr               = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}