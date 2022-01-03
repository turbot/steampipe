locals {
  pci_v321_sagemaker_common_tags = merge(local.pci_v321_common_tags, {
    service = "sagemaker"
  })
}

benchmark "pci_v321_sagemaker" {
  title         = "SageMaker"
  documentation = file("./pci_v321/docs/pci_v321_sagemaker.md")
  children = [
    control.pci_v321_sagemaker_1,
  ]
  tags = local.pci_v321_sagemaker_common_tags
}

control "pci_v321_sagemaker_1" {
  title         = "1 Amazon SageMaker notebook instances should not have direct internet access"
  description   = "This control checks whether direct internet access is disabled for an SageMaker notebook instance. To do this, it checks whether the DirectInternetAccess field is disabled for the notebook instance."
  severity      = "high"
  sql           = query.sagemaker_notebook_instance_direct_internet_access_disabled.sql
  documentation = file("./pci_v321/docs/pci_v321_sagemaker_1.md")

  tags = merge(local.pci_v321_sagemaker_common_tags, {
    pci_item_id      = "sagemaker_1"
    pci_requirements = "1.2.1,1.3.1,1.3.2,1.3.4,1.3.6"
  })
}