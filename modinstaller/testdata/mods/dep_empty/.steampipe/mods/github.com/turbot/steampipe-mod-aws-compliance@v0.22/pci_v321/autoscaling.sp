locals {
  pci_v321_autoscaling_common_tags = merge(local.pci_v321_common_tags, {
    service = "autoscaling"
  })
}

benchmark "pci_v321_autoscaling" {
  title         = "Auto Scaling"
  documentation = file("./pci_v321/docs/pci_v321_autoscaling_1.md")
  children = [
    control.pci_v321_autoscaling_1,
  ]
  tags = local.pci_v321_autoscaling_common_tags
}

control "pci_v321_autoscaling_1" {
  title         = "1 Auto Scaling groups associated with a load balancer should use health checks"
  description   = "This control checks whether your Auto Scaling groups that are associated with a load balancer are using Elastic Load Balancing health checks. PCI DSS does not require load balancing or highly available configurations. However, this check aligns with AWS best practices."
  severity      = "low"
  sql           = query.autoscaling_group_with_lb_use_health_check.sql
  documentation = file("./pci_v321/docs/pci_v321_autoscaling_1.md")

  tags = merge(local.pci_v321_autoscaling_common_tags, {
    pci_item_id      = "autoscaling_1"
    pci_requirements = "2.2"
  })
}