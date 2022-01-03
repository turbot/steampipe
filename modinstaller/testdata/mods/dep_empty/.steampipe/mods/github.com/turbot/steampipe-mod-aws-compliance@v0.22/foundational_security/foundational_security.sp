locals {
  foundational_security_common_tags = {
    aws_foundational_security = "true"
    plugin                    = "aws"
  }
}

benchmark "foundational_security" {
  title         = "AWS Foundational Security Best Practices"
  description   = "The AWS Foundational Security Best Practices standard is a set of controls that detect when your deployed accounts and resources deviate from security best practices."
  documentation = file("./foundational_security/docs/foundational_security_overview.md")
  children = [
    benchmark.foundational_security_acm,
    benchmark.foundational_security_apigateway,
    benchmark.foundational_security_autoscaling,
    benchmark.foundational_security_cloudfront,
    benchmark.foundational_security_cloudtrail,
    benchmark.foundational_security_codebuild,
    benchmark.foundational_security_config,
    benchmark.foundational_security_dms,
    benchmark.foundational_security_dynamodb,
    benchmark.foundational_security_ec2,
    benchmark.foundational_security_ecs,
    benchmark.foundational_security_efs,
    benchmark.foundational_security_elasticbeanstalk,
    benchmark.foundational_security_elb,
    benchmark.foundational_security_elbv2,
    benchmark.foundational_security_emr,
    benchmark.foundational_security_es,
    benchmark.foundational_security_guardduty,
    benchmark.foundational_security_iam,
    benchmark.foundational_security_kms,
    benchmark.foundational_security_lambda,
    benchmark.foundational_security_rds,
    benchmark.foundational_security_redshift,
    benchmark.foundational_security_s3,
    benchmark.foundational_security_sagemaker,
    benchmark.foundational_security_secretsmanager,
    benchmark.foundational_security_sns,
    benchmark.foundational_security_ssm,
    benchmark.foundational_security_sqs
  ]
  tags = local.foundational_security_common_tags
}
