locals {
  pci_v321_elbv2_common_tags = merge(local.pci_v321_common_tags, {
    service = "elbv2"
  })
}

benchmark "pci_v321_elbv2" {
  title         = "ELBV2"
  documentation = file("./pci_v321/docs/pci_v321_elbv2.md")
  children = [
    control.pci_v321_elbv2_1
  ]
  tags          = local.pci_v321_elbv2_common_tags
}

control "pci_v321_elbv2_1" {
  title         = "1 Application Load Balancer should be configured to redirect all HTTP requests to HTTPS"
  description   = "This control checks whether HTTP to HTTPS redirection is configured on all HTTP listeners of Application Load Balancers. The control fails if any of the HTTP listeners of Application Load Balancers do not have HTTP to HTTPS redirection configured."
  severity      = "medium"
  sql           = query.elb_application_lb_redirect_http_request_to_https.sql
  documentation = file("./pci_v321/docs/pci_v321_elbv2_1.md")

  tags = merge(local.pci_v321_elbv2_common_tags, {
    pci_item_id      = "elbv2_1"
    pci_requirements = "2.3,4.1"
  })
}