locals {
  foundational_security_ecs_common_tags = merge(local.foundational_security_common_tags, {
    service = "ecs"
  })
}

benchmark "foundational_security_ecs" {
  title         = "Elastic Container Service"
  documentation = file("./foundational_security/docs/foundational_security_ecs.md")
  children = [
    control.foundational_security_ecs_1,
    control.foundational_security_ecs_2
  ]
  tags          = local.foundational_security_ecs_common_tags
}

control "foundational_security_ecs_1" {
  title         = "1 Amazon ECS task definitions should have secure networking modes and user definitions"
  description   = "This control checks whether an Amazon ECS task definition that has host networking mode also has 'privileged' or 'user' container definitions. The control fails for task definitions that have host network mode and container definitions where privileged=false or is empty and user=root or is empty."
  severity      = "medium"
  sql           = query.ecs_task_definition_user_for_host_mode_check.sql
  documentation = file("./foundational_security/docs/foundational_security_ecs_1.md")

  tags = merge(local.foundational_security_ecs_common_tags, {
    foundational_security_item_id  = "ecs_1"
    foundational_security_category = "secure_access_management"
  })
}

control "foundational_security_ecs_2" {
  title         = "2 Amazon ECS services should not have public IP addresses assigned to them automatically"
  description   = "This control checks whether Amazon ECS services are configured to automatically assign public IP addresses. This control fails if AssignPublicIP is ENABLED. This control passes if AssignPublicIP is DISABLED."
  severity      = "high"
  sql           = query.ecs_service_not_publicly_accessible.sql
  #documentation = file("./foundational_security/docs/foundational_security_ecs_2.md")

  tags = merge(local.foundational_security_ecs_common_tags, {
    foundational_security_item_id  = "ecs_2"
    foundational_security_category = "resources_not_publicly_accessible"
  })
}
