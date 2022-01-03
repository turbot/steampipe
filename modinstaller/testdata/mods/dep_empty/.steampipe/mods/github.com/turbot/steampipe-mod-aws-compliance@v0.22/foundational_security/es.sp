locals {
  foundational_security_es_common_tags = merge(local.foundational_security_common_tags, {
    service = "es"
  })
}

benchmark "foundational_security_es" {
  title         = "Elasticsearch"
  documentation = file("./foundational_security/docs/foundational_security_es.md")
  children = [
    control.foundational_security_es_1,
    control.foundational_security_es_2,
    control.foundational_security_es_3,
    control.foundational_security_es_4,
    control.foundational_security_es_5,
    control.foundational_security_es_6,
    control.foundational_security_es_7,
    control.foundational_security_es_8
  ]
  tags          = local.foundational_security_es_common_tags
}

control "foundational_security_es_1" {
  title         = "1 Elasticsearch domains should have encryption at-rest enabled"
  description   = "This control checks whether Amazon Elasticsearch Service (Amazon ES) domains have encryption at rest configuration enabled. The check fails if encryption at rest is not enabled."
  severity      = "medium"
  sql           = query.es_domain_encryption_at_rest_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_es_1.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_1"
    foundational_security_category = "encryption_of_data_at_rest"
  })
}

control "foundational_security_es_2" {
  title         = "2 Amazon Elasticsearch Service domains should be in a VPC"
  description   = "This control checks whether Amazon Elasticsearch Service domains are in a VPC. It does not evaluate the VPC subnet routing configuration to determine public access. You should ensure that Amazon ES domains are not attached to public subnets."
  severity      = "critical"
  sql           = query.es_domain_in_vpc.sql
  documentation = file("./foundational_security/docs/foundational_security_es_2.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_2"
    foundational_security_category = "resources_within_vpc"
  })
}

control "foundational_security_es_3" {
  title         = "3 Amazon Elasticsearch Service domains should encrypt data sent between nodes"
  description   = "This control checks whether Amazon ES domains have node-to-node encryption enabled."
  severity      = "medium"
  sql           = query.es_domain_node_to_node_encryption_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_es_3.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_3"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}

control "foundational_security_es_4" {
  title         = "4 Elasticsearch domain error logging to CloudWatch Logs should be enabled"
  description   = "This control checks whether Elasticsearch domains are configured to send error logs to CloudWatch Logs."
  severity      = "medium"
  sql           = query.es_domain_error_logging_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_es_4.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_4"
    foundational_security_category = "logging"
  })
}

control "foundational_security_es_5" {
  title         = "5 Elasticsearch domains should have audit logging enabled"
  description   = "This control checks whether Elasticsearch domains have audit logging enabled. This control fails if an Elasticsearch domain does not have audit logging enabled."
  severity      = "medium"
  sql           = query.es_domain_audit_logging_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_es_4.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_5"
    foundational_security_category = "logging"
  })
}

control "foundational_security_es_6" {
  title         = "6 Elasticsearch domains should have at least three data nodes"
  description   = "This control checks whether Elasticsearch domains are configured with at least three data nodes and zoneAwarenessEnabled is true."
  severity      = "medium"
  sql           = query.es_domain_data_nodes_min_3.sql
  documentation = file("./foundational_security/docs/foundational_security_es_6.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_6"
    foundational_security_category = "high_availability"
  })
}

control "foundational_security_es_7" {
  title         = "7 Elasticsearch domains should be configured with at least three dedicated master nodes"
  description   = "This control checks whether Elasticsearch domains are configured with at least three dedicated master nodes. This control fails if the domain does not use dedicated master nodes. This control passes if Elasticsearch domains have five dedicated master nodes. However, using more than three master nodes might be unnecessary to mitigate the availability risk, and will result in additional cost."
  severity      = "medium"
  sql           = query.es_domain_dedicated_master_nodes_min_3.sql
  documentation = file("./foundational_security/docs/foundational_security_es_7.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_7"
    foundational_security_category = "high_availability"
  })
}

control "foundational_security_es_8" {
  title         = "8 Connections to Elasticsearch domains should be encrypted using TLS 1.2"
  description   = "This control checks whether connections to Elasticsearch domains are required to use TLS 1.2. The check fails if the Elasticsearch domain TLSSecurityPolicy is not Policy-Min-TLS-1-2-2019-07."
  severity      = "medium"
  sql           = query.es_domain_encrypted_using_tls_1_2.sql
  documentation = file("./foundational_security/docs/foundational_security_es_8.md")

  tags = merge(local.foundational_security_es_common_tags, {
    foundational_security_item_id  = "es_8"
    foundational_security_category = "encryption_of_data_in_transit"
  })
}