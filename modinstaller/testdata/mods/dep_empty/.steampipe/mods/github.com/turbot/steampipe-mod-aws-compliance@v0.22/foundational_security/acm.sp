locals {
  foundational_security_acm_common_tags = merge(local.foundational_security_common_tags, {
    service = "acm"
  })
}

benchmark "foundational_security_acm" {
  title         = "ACM"
  documentation = file("./foundational_security/docs/foundational_security_acm.md")
  children = [
    control.foundational_security_acm_1
  ]
  tags          = local.foundational_security_acm_common_tags
}

control "foundational_security_acm_1" {
  title         = "1 Imported ACM certificates should be renewed after a specified time period"
  description   = "This control checks whether ACM certificates in your account are marked for expiration within 30 days. It checks both imported certificates and certificates provided by AWS Certificate Manager."
  severity      = "medium"
  sql           = query.acm_certificate_expires_30_days.sql
  documentation = file("./foundational_security/docs/foundational_security_acm_1.md")

  tags = merge(local.foundational_security_acm_common_tags, {
    foundational_security_item_id  = "acm_1"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}