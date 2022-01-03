locals {
  conformance_pack_redshift_common_tags = {
    service = "redshift"
  }
}

control "redshift_cluster_encryption_in_transit_enabled" {
  title       = "Redshift cluster encryption in transit should be enabled"
  description = "Ensure that your Amazon Redshift clusters require TLS/SSL encryption to connect to SQL clients."
  sql         = query.redshift_cluster_encryption_in_transit_enabled.sql

  tags = merge(local.conformance_pack_redshift_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "redshift_cluster_encryption_logging_enabled" {
  title       = "Redshift cluster audit logging and encryption should be enabled"
  description = "To protect data at rest, ensure that encryption is enabled for your Amazon Redshift clusters. You must also ensure that required configurations are deployed on Amazon Redshift clusters. The audit logging should be enabled to provide information about connections and user activities in the database."
  sql         = query.redshift_cluster_encryption_logging_enabled.sql

  tags = merge(local.conformance_pack_redshift_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "redshift_cluster_prohibit_public_access" {
  title       = "Redshift clusters should prohibit public access"
  description = "Manage access to resources in the AWS Cloud by ensuring that Amazon Redshift clusters are not public."
  sql         = query.redshift_cluster_prohibit_public_access.sql

  tags = merge(local.conformance_pack_redshift_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "redshift_cluster_automatic_snapshots_min_7_days" {
  title       = "Amazon Redshift clusters should have automatic snapshots enabled"
  description = "This control checks whether Amazon Redshift clusters have automated snapshots enabled. It also checks whether the snapshot retention period is greater than or equal to seven."
  sql         = query.redshift_cluster_automatic_snapshots_min_7_days.sql

  tags = merge(local.conformance_pack_redshift_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    sco_2              = "true"
  })
}

control "redshift_cluster_kms_enabled" {
  title       = "Amazon Redshift clusters should be encrypted with KMS"
  description = "Ensure if Amazon Redshift clusters are using a specified AWS Key Management Service (AWS KMS) key for encryption. The rule is complaint if encryption is enabled and the cluster is encrypted with the key provided in the kmsKeyArn parameter. The rule is non complaint if the cluster is not encrypted or encrypted with another key."
  sql         = query.redshift_cluster_kms_enabled.sql

  tags = merge(local.conformance_pack_redshift_common_tags, {
    rbi_cyber_security = "true"
  })
}

control "redshift_cluster_maintenance_settings_check" {
  title       = "Amazon Redshift should have required maintenance settings"
  description = "Ensure whether Amazon Redshift clusters have the specified maintenance settings. Redshift clusters `allowVersionUpgrade` should be set to `true` and `automatedSnapshotRetentionPeriod` should be greater than 7."
  sql         = query.redshift_cluster_maintenance_settings_check.sql

  tags = merge(local.conformance_pack_redshift_common_tags, {
    rbi_cyber_security = "true"
  })
}