locals {
  foundational_security_sns_common_tags = merge(local.foundational_security_common_tags, {
    service = "sns"
  })
}

benchmark "foundational_security_sns" {
  title         = "SNS"
  documentation = file("./foundational_security/docs/foundational_security_sns.md")
  children = [
    control.foundational_security_sns_1
  ]
  tags          = local.foundational_security_sns_common_tags
}

control "foundational_security_sns_1" {
  title         = "1 SNS topics should be encrypted at rest using AWS KMS"
  description   = "This control checks whether an SNS topic is encrypted at rest using AWS KMS."
  severity      = "medium"
  sql           = query.sns_topic_encrypted_at_rest.sql
  documentation = file("./foundational_security/docs/foundational_security_sns_1.md")

  tags = merge(local.foundational_security_sns_common_tags, {
    foundational_security_item_id  = "sns_1"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}