locals {
  foundational_security_elb_common_tags = merge(local.foundational_security_common_tags, {
    service = "elb"
  })
}

benchmark "foundational_security_elb" {
  title         = "ELB"
  documentation = file("./foundational_security/docs/foundational_security_elb.md")
  children = [
    control.foundational_security_elb_3,
    control.foundational_security_elb_4,
    control.foundational_security_elb_5,
    control.foundational_security_elb_6,
    control.foundational_security_elb_7
  ]
  tags          = local.foundational_security_elb_common_tags
}

control "foundational_security_elb_3" {
  title         = "3 Classic Load Balancer listeners should be configured with HTTPS or TLS termination"
  description   = "This control checks whether your Classic Load Balancer listeners are configured with HTTPS or TLS protocol for front-end (client to load balancer) connections. The control is applicable if a Classic Load Balancer has listeners. If your Classic Load Balancer does not have a listener configured, then the control does not report any findings. The control passes if the Classic Load Balancer listeners are configured with TLS or HTTPS for front-end connections. The control fails if the listener is not configured with TLS or HTTPS for front-end connections."
  severity      = "medium"
  sql           = query.elb_classic_lb_use_tls_https_listeners.sql
  documentation = file("./foundational_security/docs/foundational_security_elb_3.md")

  tags = merge(local.foundational_security_elb_common_tags, {
    foundational_security_item_id  = "elb_3"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}

control "foundational_security_elb_4" {
  title         = "4 Application load balancers should be configured to drop HTTP headers"
  description   = "This control evaluates AWS Application Load Balancers (ALB) to ensure they are configured to drop invalid HTTP headers. The control fails if the value of routing.http.drop_invalid_header_fields.enabled is set to false. By default, ALBs are not configured to drop invalid HTTP header values. Removing these header values prevents HTTP desync attacks."
  severity      = "medium"
  sql           = query.elb_application_lb_drop_http_headers.sql
  documentation = file("./foundational_security/docs/foundational_security_elb_4.md")

  tags = merge(local.foundational_security_elb_common_tags, {
    foundational_security_item_id  = "elb_4"
    foundational_security_category = "network_security"
  })
}

control "foundational_security_elb_5" {
  title         = "5 Application and Classic Load Balancers logging should be enabled"
  description   = "This control checks whether the Application Load Balancer and the Classic Load Balancer have logging enabled. The control fails if access_logs.s3.enabled is false."
  severity      = "medium"
  sql           = query.elb_application_classic_lb_logging_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_elb_5.md")

  tags = merge(local.foundational_security_elb_common_tags, {
    foundational_security_item_id  = "elb_5"
    foundational_security_category = "logging"
  })
}

control "foundational_security_elb_6" {
  title         = "6 Application Load Balancer deletion protection should be enabled"
  description   = "This control checks whether an Application Load Balancer has deletion protection enabled. The control fails if deletion protection is not configured. Enable deletion protection to protect your Application Load Balancer from deletion."
  severity      = "medium"
  sql           = query.elb_application_lb_deletion_protection_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_elb_6.md")

  tags = merge(local.foundational_security_elb_common_tags, {
    foundational_security_item_id  = "elb_6"
    foundational_security_category = "high_availability"
  })
}

control "foundational_security_elb_7" {
  title         = "7 Classic Load Balancers should have connection draining enabled"
  description   = "This control checks whether Classic Load Balancers have connection draining enabled."
  severity      = "medium"
  sql           = query.ec2_classic_lb_connection_draining_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_elb_7.md")

  tags = merge(local.foundational_security_elb_common_tags, {
    foundational_security_item_id  = "elb_7"
    foundational_security_category = "resilience"
  })
}
