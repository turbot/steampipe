locals {
  foundational_security_sqs_common_tags = merge(local.foundational_security_common_tags, {
    service = "sqs"
  })
}

benchmark "foundational_security_sqs" {
  title         = "SQS"
  documentation = file("./foundational_security/docs/foundational_security_sqs.md")
  children = [
    control.foundational_security_sqs_1
  ]
  tags          = local.foundational_security_sqs_common_tags
}

control "foundational_security_sqs_1" {
  title         = "1 Amazon SQS queues should be encrypted at rest"
  description   = "This control checks whether Amazon SQS queues are encrypted at rest."
  severity      = "medium"
  sql           = query.sqs_queue_encrypted_at_rest.sql
  documentation = file("./foundational_security/docs/foundational_security_sqs_1.md")

  tags = merge(local.foundational_security_sqs_common_tags, {
    foundational_security_item_id  = "sqs_1"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}