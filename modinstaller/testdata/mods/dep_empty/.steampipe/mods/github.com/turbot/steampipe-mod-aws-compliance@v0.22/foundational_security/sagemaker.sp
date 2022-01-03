locals {
  foundational_security_sagemaker_common_tags = merge(local.foundational_security_common_tags, {
    service = "sagemaker"
  })
}

benchmark "foundational_security_sagemaker" {
  title         = "SageMaker"
  documentation = file("./foundational_security/docs/foundational_security_sagemaker.md")
  children = [
    control.foundational_security_sagemaker_1
  ]
  tags          = local.foundational_security_sagemaker_common_tags
}

control "foundational_security_sagemaker_1" {
  title         = "1 SageMaker notebook instances should not have direct internet access"
  description   = "This control checks whether direct internet access is disabled for an SageMaker notebook instance. To do this, it checks whether the DirectInternetAccess field is disabled for the notebook instance. If you configure your SageMaker instance without a VPC, then by default direct internet access is enabled on your instance. You should configure your instance with a VPC and change the default setting to Disable — Access the internet through a VPC."
  severity      = "high"
  sql           = query.sagemaker_notebook_instance_direct_internet_access_disabled.sql
  documentation = file("./foundational_security/docs/foundational_security_sagemaker_1.md")

  tags = merge(local.foundational_security_sagemaker_common_tags, {
    foundational_security_item_id  = "sagemaker_1"
    foundational_security_category = "secure_network_configuration"
  })
}