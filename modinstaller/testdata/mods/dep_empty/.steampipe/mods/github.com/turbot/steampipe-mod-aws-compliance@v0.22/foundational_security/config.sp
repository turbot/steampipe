locals {
  foundational_security_config_common_tags = merge(local.foundational_security_common_tags, {
    service = "config"
  })
}

benchmark "foundational_security_config" {
  title         = "Config"
  documentation = file("./foundational_security/docs/foundational_security_config.md")
  children = [
    control.foundational_security_config_1
  ]
  tags          = local.foundational_security_config_common_tags
}

control "foundational_security_config_1" {
  title         = "1 AWS Config should be enabled"
  description   = "This control checks whether AWS Config is enabled in the account for the local Region and is recording all resources. The AWS Config service performs configuration management of supported AWS resources in your account and delivers log files to you. The recorded information includes the configuration item (AWS resource), relationships between configuration items, and any configuration changes between resources."
  severity      = "medium"
  sql           = query.config_enabled_all_regions.sql
  documentation = file("./foundational_security/docs/foundational_security_config_1.md")

  tags = merge(local.foundational_security_config_common_tags, {
    foundational_security_item_id  = "config_1"
    foundational_security_category = "inventory"
  })
}