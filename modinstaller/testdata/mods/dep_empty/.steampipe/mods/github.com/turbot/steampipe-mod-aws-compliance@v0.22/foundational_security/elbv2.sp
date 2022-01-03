locals {
  foundational_security_elbv2_common_tags = merge(local.foundational_security_common_tags, {
    service = "elbv2"
  })
}

benchmark "foundational_security_elbv2" {
  title         = "ELBv2"
  documentation = file("./foundational_security/docs/foundational_security_elbv2.md")
  children = [
    control.foundational_security_elbv2_1
  ]
  tags          = local.foundational_security_elbv2_common_tags
}

control "foundational_security_elbv2_1" {
  title         = "1 Application Load Balancer should be configured to redirect all HTTP requests to HTTPS"
  description   = "This control checks whether HTTP to HTTPS redirection is configured on all HTTP listeners of Application Load Balancers. The check fails if one or more HTTP listeners of Application Load Balancers do not have HTTP to HTTPS redirection configured."
  severity      = "medium"
  sql           = query.elb_application_lb_redirect_http_request_to_https.sql
  documentation = file("./foundational_security/docs/foundational_security_elbv2_1.md")

  tags = merge(local.foundational_security_elbv2_common_tags, {
    foundational_security_item_id  = "elbv2_1"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}