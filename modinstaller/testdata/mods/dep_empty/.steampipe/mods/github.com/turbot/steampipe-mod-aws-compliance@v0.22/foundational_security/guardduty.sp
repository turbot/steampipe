locals {
  foundational_security_guardduty_common_tags = merge(local.foundational_security_common_tags, {
    service = "guardduty"
  })
}

benchmark "foundational_security_guardduty" {
  title         = "GuardDuty"
  documentation = file("./foundational_security/docs/foundational_security_guardduty.md")
  children = [
    control.foundational_security_guardduty_1
  ]
  tags          = local.foundational_security_guardduty_common_tags
}

control "foundational_security_guardduty_1" {
  title         = "1 GuardDuty should be enabled"
  description   = "This control checks whether Amazon GuardDuty is enabled in your GuardDuty account and Region. It is highly recommended that you enable GuardDuty in all supported AWS Regions. Doing so allows GuardDuty to generate findings about unauthorized or unusual activity, even in Regions that you do not actively use. This also allows GuardDuty to monitor CloudTrail events for global AWS services such as IAM."
  severity      = "high"
  sql           = query.guardduty_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_guardduty_1.md")

  tags = merge(local.foundational_security_guardduty_common_tags, {
    foundational_security_item_id  = "guardduty_1"
    foundational_security_category = "detection_services"
  })
}