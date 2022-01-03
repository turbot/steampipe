locals {
  conformance_pack_dms_common_tags = {
    service = "dms"
  }
}

control "dms_replication_instance_not_publicly_accessible" {
  title       = "DMS replication instances should not be publicly accessible"
  description = "Manage access to the AWS Cloud by ensuring DMS replication instances cannot be publicly accessed."
  sql         = query.dms_replication_instance_not_publicly_accessible.sql

  tags = merge(local.conformance_pack_dms_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}