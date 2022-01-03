locals {
  conformance_pack_apigateway_common_tags = {
    service = "apigateway"
  }
}

control "apigateway_stage_cache_encryption_at_rest_enabled" {
  title       = "API Gateway stage cache encryption at rest should be enabled"
  description = "To help protect data at rest, ensure encryption is enabled for your API Gateway stage's cache."
  sql         = query.apigateway_stage_cache_encryption_at_rest_enabled.sql

  tags = merge(local.conformance_pack_apigateway_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "apigateway_stage_logging_enabled" {
  title       = "API Gateway stage logging should be enabled"
  description = "API Gateway logging displays detailed views of users who accessed the API and the way they accessed the API."
  sql         = query.apigateway_stage_logging_enabled.sql

  tags = merge(local.conformance_pack_apigateway_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "apigateway_rest_api_stage_use_ssl_certificate" {
  title       = "API Gateway stage should uses SSL certificate"
  description = "Ensure if a REST API stage uses a Secure Sockets Layer (SSL) certificate. This rule is complaint if the REST API stage does not have an associated SSL certificate."
  sql         = query.apigateway_rest_api_stage_use_ssl_certificate.sql

  tags = merge(local.conformance_pack_apigateway_common_tags, {
    rbi_cyber_security = "true"
  })
}

control "apigateway_stage_use_waf_web_acl" {
  title       = "API Gateway stage should be associated with waf"
  description = "Ensure if an Amazon API Gateway API stage is using a WAF Web ACL. This rule is non complaint if an AWS WAF Web ACL is not used."
  sql         = query.apigateway_stage_use_waf_web_acl.sql

  tags = merge(local.conformance_pack_apigateway_common_tags, {
    rbi_cyber_security = "true"
  })
}