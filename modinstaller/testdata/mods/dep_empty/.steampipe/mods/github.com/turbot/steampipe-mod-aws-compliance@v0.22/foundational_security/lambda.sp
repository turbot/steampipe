locals {
  foundational_security_lambda_common_tags = merge(local.foundational_security_common_tags, {
    service = "lambda"
  })
}

benchmark "foundational_security_lambda" {
  title         = "Lambda"
  documentation = file("./foundational_security/docs/foundational_security_lambda.md")
  children = [
    control.foundational_security_lambda_1,
    control.foundational_security_lambda_2,
    control.foundational_security_lambda_4
  ]
  tags          = local.foundational_security_lambda_common_tags
}

control "foundational_security_lambda_1" {
  title         = "1 Lambda function policies should prohibit public access"
  description   = "This control checks whether the Lambda function resource-based policy prohibits public access outside of your account. The Lambda function should not be publicly accessible, as this may allow unintended access to your code stored in the function."
  severity      = "critical"
  sql           = query.lambda_function_restrict_public_access.sql
  documentation = file("./foundational_security/docs/foundational_security_lambda_1.md")

  tags = merge(local.foundational_security_lambda_common_tags, {
    foundational_security_item_id  = "lambda_1"
    #foundational_security_category = "secure_network_configuration"
  })
}

control "foundational_security_lambda_2" {
  title         = "2 Lambda functions should use latest runtimes"
  description   = "This control checks that the Lambda function settings for runtimes match the expected values set for the latest runtimes for each supported language. This control checks for the following runtimes: nodejs14.x, nodejs12.x, nodejs10.x, python3.8, python3.7, python3.6, ruby2.7, ruby2.5,java11, java8, go1.x, dotnetcore3.1, dotnetcore2.1."
  severity      = "medium"
  sql           = query.lambda_function_use_latest_runtime.sql
  documentation = file("./foundational_security/docs/foundational_security_lambda_2.md")

  tags = merge(local.foundational_security_lambda_common_tags, {
    foundational_security_item_id  = "lambda_2"
    #foundational_security_category = "secure_development"
  })
}

control "foundational_security_lambda_4" {
  title         = "4 Lambda functions should have a dead-letter queue configured"
  description   = "This control checks whether a Lambda function is configured with a dead-letter queue. The control fails if the Lambda function is not configured with a dead-letter queue."
  severity      = "medium"
  sql           = query.lambda_function_dead_letter_queue_configured.sql
  documentation = file("./foundational_security/docs/foundational_security_lambda_4.md")

  tags = merge(local.foundational_security_lambda_common_tags, {
    foundational_security_item_id  = "lambda_4"
    foundational_security_category = "logging"
  })
}