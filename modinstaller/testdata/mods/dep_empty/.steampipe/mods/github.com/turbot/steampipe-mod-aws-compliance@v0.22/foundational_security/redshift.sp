locals {
  foundational_security_redshift_common_tags = merge(local.foundational_security_common_tags, {
    service = "redshift"
  })
}

benchmark "foundational_security_redshift" {
  title         = "Redshift"
  documentation = file("./foundational_security/docs/foundational_security_redshift.md")
  children = [
    control.foundational_security_redshift_1,
    control.foundational_security_redshift_2,
    control.foundational_security_redshift_3,
    control.foundational_security_redshift_4,
    control.foundational_security_redshift_6,
    control.foundational_security_redshift_7
  ]
  tags          = local.foundational_security_redshift_common_tags
}

control "foundational_security_redshift_1" {
  title         = "1 Amazon Redshift clusters should prohibit public access"
  description   = "This control checks whether Amazon Redshift clusters are publicly accessible. It evaluates the PubliclyAccessible field in the cluster configuration item. The PubliclyAccessible attribute of the Amazon Redshift cluster configuration indicates whether the cluster is publicly accessible. When the cluster is configured with PubliclyAccessible set to true, it is an Internet-facing instance that has a publicly resolvable DNS name, which resolves to a public IP address. When the cluster is not publicly accessible, it is an internal instance with a DNS name that resolves to a private IP address. Unless you intend for your cluster to be publicly accessible, the cluster should not be configured with PubliclyAccessible set to true."
  severity      = "critical"
  sql           = query.redshift_cluster_prohibit_public_access.sql
  documentation = file("./foundational_security/docs/foundational_security_redshift_1.md")

  tags = merge(local.foundational_security_redshift_common_tags, {
    foundational_security_item_id  = "redshift_1"
    foundational_security_category = "resources_not_publicly_accessible"
  })
}

control "foundational_security_redshift_2" {
  title         = "2 Connections to Amazon Redshift clusters should be encrypted in transit"
  description   = "This control checks whether connections to Amazon Redshift clusters are required to use encryption in transit. The check fails if the Amazon Redshift cluster parameter require_SSL is not set to 1. TLS can be used to help prevent potential attackers from using person-in-the-middle or similar attacks to eavesdrop on or manipulate network traffic. Only encrypted connections over TLS should be allowed. Encrypting data in transit can affect performance. You should test your application with this feature to understand the performance profile and the impact of TLS."
  severity      = "medium"
  sql           = query.redshift_cluster_encryption_in_transit_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_redshift_2.md")

  tags = merge(local.foundational_security_redshift_common_tags, {
    foundational_security_item_id  = "redshift_2"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}

control "foundational_security_redshift_3" {
  title         = "3 Amazon Redshift clusters should have automatic snapshots enabled"
  description   = "This control checks whether Amazon Redshift clusters have automated snapshots enabled. It also checks whether the snapshot retention period is greater than or equal to seven."
  severity      = "medium"
  sql           = query.redshift_cluster_automatic_snapshots_min_7_days.sql
  documentation = file("./foundational_security/docs/foundational_security_redshift_3.md")

  tags = merge(local.foundational_security_redshift_common_tags, {
    foundational_security_item_id  = "redshift_3"
    foundational_security_category = "backups_enabled"
  })
}

control "foundational_security_redshift_4" {
  title         = "4 Amazon Redshift clusters should have audit logging enabled"
  description   = "This control checks whether an Amazon Redshift cluster has audit logging enabled."
  severity      = "medium"
  sql           = query.redshift_cluster_automatic_snapshots_min_7_days.sql
  documentation = file("./foundational_security/docs/foundational_security_redshift_4.md")

  tags = merge(local.foundational_security_redshift_common_tags, {
    foundational_security_item_id  = "redshift_4"
    foundational_security_category = "logging"
  })
}

control "foundational_security_redshift_6" {
  title         = "6 Amazon Redshift should have automatic upgrades to major versions enabled"
  description   = "This control checks whether automatic major version upgrades are enabled for the Amazon Redshift cluster."
  severity      = "medium"
  sql           = query.redshift_cluster_automatic_upgrade_major_versions_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_redshift_6.md")

  tags = merge(local.foundational_security_redshift_common_tags, {
    foundational_security_item_id  = "redshift_6"
    foundational_security_category = "vulnerability_and_patch_management"
  })
}

control "foundational_security_redshift_7" {
  title         = "7 Amazon Redshift clusters should use enhanced VPC routing"
  description   = "This control checks whether an Amazon Redshift cluster has EnhancedVpcRouting enabled."
  severity      = "high"
  sql           = query.redshift_cluster_enhanced_vpc_routing_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_redshift_7.md")

  tags = merge(local.foundational_security_redshift_common_tags, {
    foundational_security_item_id  = "redshift_7"
    foundational_security_category = "api_private_access"
  })
}