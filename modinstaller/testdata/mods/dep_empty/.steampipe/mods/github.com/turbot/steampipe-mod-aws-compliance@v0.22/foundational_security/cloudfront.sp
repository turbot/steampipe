locals {
  foundational_security_cloudfront_common_tags = merge(local.foundational_security_common_tags, {
    service = "cloudfront"
  })
}

benchmark "foundational_security_cloudfront" {
  title         = "CloudFront"
  documentation = file("./foundational_security/docs/foundational_security_cloudfront.md")
  children = [
    control.foundational_security_cloudfront_1,
    control.foundational_security_cloudfront_2,
    control.foundational_security_cloudfront_3,
    control.foundational_security_cloudfront_4,
    control.foundational_security_cloudfront_5,
    control.foundational_security_cloudfront_6
  ]
  tags          = local.foundational_security_cloudfront_common_tags
}

control "foundational_security_cloudfront_1" {
  title         = "1 CloudFront distributions should have a default root object configured"
  description   = "This control checks whether an Amazon CloudFront distribution is configured to return a specific object that is the default root object. The control fails if the CloudFront distribution does not have a default root object configured."
  severity      = "critical"
  sql           = query.cloudfront_distribution_default_root_object_configured.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudfront_1.md")

  tags = merge(local.foundational_security_cloudfront_common_tags, {
    foundational_security_item_id  = "cloudfront_1"
    foundational_security_category = "resources_not_publicly_accessible"
  })
}

control "foundational_security_cloudfront_2" {
  title         = "2 CloudFront distributions should have origin access identity enabled"
  description   = "This control checks whether an Amazon CloudFront distribution with Amazon S3 Origin type has Origin Access Identity (OAI) configured. The control fails if OAI is not configured."
  severity      = "medium"
  sql           = query.cloudfront_distribution_origin_access_identity_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudfront_2.md")

  tags = merge(local.foundational_security_cloudfront_common_tags, {
    foundational_security_item_id  = "cloudfront_2"
    foundational_security_category = "resource_policy_configuration"
  })
}

control "foundational_security_cloudfront_3" {
  title         = "3 CloudFront distributions should require encryption in transit"
  description   = "This control checks whether an Amazon CloudFront distribution requires viewers to use HTTPS directly or whether it uses redirection. The control fails if ViewerProtocolPolicy is set to allow-all for defaultCacheBehavior or for cacheBehaviors."
  severity      = "medium"
  sql           = query.cloudfront_distribution_encryption_in_transit_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudfront_3.md")

  tags = merge(local.foundational_security_cloudfront_common_tags, {
    foundational_security_item_id  = "cloudfront_3"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}

control "foundational_security_cloudfront_4" {
  title         = "4 CloudFront distributions should have origin failover configured"
  description   = "This control checks whether an Amazon CloudFront distribution is configured with an origin group that has two or more origins. CloudFront origin failover can increase availability. Origin failover automatically redirects traffic to a secondary origin if the primary origin is unavailable or if it returns specific HTTP response status codes."
  severity      = "low"
  sql           = query.cloudfront_distribution_configured_with_origin_failover.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudfront_4.md")

  tags = merge(local.foundational_security_cloudfront_common_tags, {
    foundational_security_item_id  = "cloudfront_4"
    foundational_security_category = "high_availability"
  })
}

control "foundational_security_cloudfront_5" {
  title         = "5 CloudFront distributions should have logging enabled"
  description   = "This control checks whether server access logging is enabled on CloudFront distributions. The control fails if access logging is not enabled for a distribution."
  severity      = "medium"
  sql           = query.cloudfront_distribution_logging_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudfront_5.md")

  tags = merge(local.foundational_security_cloudfront_common_tags, {
    foundational_security_item_id  = "cloudfront_5"
    foundational_security_category = "logging"
  })
}

control "foundational_security_cloudfront_6" {
  title         = "6 CloudFront distributions should have AWS WAF enabled"
  description   = "This control checks whether CloudFront distributions are associated with either AWS WAF or AWS WAFv2 web ACLs. The control fails if the distribution is not associated with a web ACL."
  severity      = "medium"
  sql           = query.cloudfront_distribution_waf_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_cloudfront_6.md")

  tags = merge(local.foundational_security_cloudfront_common_tags, {
    foundational_security_item_id  = "cloudfront_6"
    foundational_security_category = "protective_services"
  })
}