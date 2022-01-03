locals {
  audit_manager_control_tower_common_tags = {
    audit_manager_control_tower = "true"
    plugin                      = "aws"
  }
}

benchmark "audit_manager_control_tower" {
  title         = "AWS Audit Manager Control Tower Guardrails"
  description   = "AWS Control Tower is a service that enables you to enforce and manage governance rules for security, operations, and compliance at scale across all your organizations and accounts in the AWS Cloud."
  documentation = file("./audit_manager_control_tower/docs/control_tower_overview.md")
  children = [
    benchmark.audit_manager_control_tower_ebs_checks,
    benchmark.audit_manager_control_tower_disallow_internet_connection,
    benchmark.audit_manager_control_tower_multi_factor_authentication,
    benchmark.audit_manager_control_tower_disallow_public_access,
    benchmark.audit_manager_control_tower_disallow_instances
  ]

  tags = local.audit_manager_control_tower_common_tags
}
