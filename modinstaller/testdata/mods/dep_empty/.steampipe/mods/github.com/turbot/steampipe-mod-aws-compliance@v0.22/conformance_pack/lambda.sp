locals {
  conformance_pack_lambda_common_tags = {
    service = "lambda"
  }
}

control "lambda_function_dead_letter_queue_configured" {
  title       = "Lambda functions should be configured with a dead-letter queue"
  description = "Enable this rule to help notify the appropriate personnel through Amazon Simple Queue Service (Amazon SQS) or Amazon Simple Notification Service (Amazon SNS) when a function has failed."
  sql         = query.lambda_function_dead_letter_queue_configured.sql

  tags = merge(local.conformance_pack_lambda_common_tags, {
    hipaa    = "true"
    nist_csf = "true"
    soc_2    = "true"
  })
}

control "lambda_function_in_vpc" {
  title       = "Lambda functions should be in a VPC"
  description = "Deploy AWS Lambda functions within an Amazon Virtual Private Cloud (Amazon VPC) for a secure communication between a function and other services within the Amazon VPC."
  sql         = query.lambda_function_in_vpc.sql

  tags = merge(local.conformance_pack_lambda_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "lambda_function_restrict_public_access" {
  title       = "Lambda functions should restrict public access"
  description = "Manage access to resources in the AWS Cloud by ensuring AWS Lambda functions cannot be publicly accessed."
  sql         = query.lambda_function_restrict_public_access.sql

  tags = merge(local.conformance_pack_lambda_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "lambda_function_concurrent_execution_limit_configured" {
  title       = "Lambda functions concurrent execution limit configured"
  description = "Checks whether the AWS Lambda function is configured with function-level concurrent execution limit. The control is non complaint if the Lambda function is not configured with function-level concurrent execution limit."
  sql         = query.lambda_function_concurrent_execution_limit_configured.sql

  tags = merge(local.conformance_pack_lambda_common_tags, {
    nist_csf = "true"
    soc_2    = "true"
  })
}
