locals {
  foundational_security_autoscaling_common_tags = merge(local.foundational_security_common_tags, {
    service = "autoscaling"
  })
}

benchmark "foundational_security_autoscaling" {
  title         = "Auto Scaling"
  documentation = file("./foundational_security/docs/foundational_security_autoscaling.md")
  children = [
    control.foundational_security_autoscaling_1
  ]
  tags          = local.foundational_security_autoscaling_common_tags
}

control "foundational_security_autoscaling_1" {
  title         = "1 Auto Scaling groups associated with a load balancer should use load balancer health checks"
  description   = "This control checks whether your Auto Scaling groups that are associated with a load balancer are using Elastic Load Balancing health checks. This ensures that the group can determine an instance's health based on additional tests provided by the load balancer. Using Elastic Load Balancing health checks can help support the availability of applications that use EC2 Auto Scaling groups."
  severity      = "low"
  sql           = query.autoscaling_group_with_lb_use_health_check.sql
  documentation = file("./foundational_security/docs/foundational_security_autoscaling_1.md")

  tags = merge(local.foundational_security_autoscaling_common_tags, {
    foundational_security_item_id  = "autoscaling_1"
    foundational_security_category = "inventory"
  })
}