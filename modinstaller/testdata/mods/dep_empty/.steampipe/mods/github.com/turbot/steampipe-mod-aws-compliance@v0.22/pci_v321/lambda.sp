locals {
  pci_v321_lambda_common_tags = merge(local.pci_v321_common_tags, {
    service = "lambda"
  })
}

benchmark "pci_v321_lambda" {
  title         = "Lambda"
  documentation = file("./pci_v321/docs/pci_v321_lambda.md")
  children = [
    control.pci_v321_lambda_1,
    control.pci_v321_lambda_2
  ]
  tags = local.pci_v321_lambda_common_tags
}

control "pci_v321_lambda_1" {
  title         = "1 Lambda functions should prohibit public access"
  description   = "This control checks whether the Lambda function resource-based policy prohibits public access. It does not check for access to the Lambda function by internal principals, such as IAM roles. You should ensure that access to the Lambda function is restricted to authorized principals only by using least privilege Lambda resource-based policies."
  severity      = "critical"
  sql           = query.lambda_function_restrict_public_access.sql
  documentation = file("./pci_v321/docs/pci_v321_lambda_1.md")

  tags = merge(local.pci_v321_lambda_common_tags, {
    pci_item_id      = "lambda_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,7.2.1"
  })
}

control "pci_v321_lambda_2" {
  title         = "2 Lambda functions should be in a VPC"
  description   = "This control checks whether a Lambda function is in a VPC. It does not evaluate the VPC subnet routing configuration to determine public reachability."
  severity      = "critical"
  sql           = query.lambda_function_in_vpc.sql
  documentation = file("./pci_v321/docs/pci_v321_lambda_2.md")

  tags = merge(local.pci_v321_lambda_common_tags, {
    pci_item_id      = "lambda_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4"
  })
}