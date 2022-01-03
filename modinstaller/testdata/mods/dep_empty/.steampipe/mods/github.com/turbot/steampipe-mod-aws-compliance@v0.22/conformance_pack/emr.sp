locals {
  conformance_pack_emr_common_tags = {
    service = "emr"
  }
}

control "emr_cluster_kerberos_enabled" {
  title       = "EMR cluster Kerberos should be enabled"
  description = "The access permissions and authorizations can be managed and incorporated with the principles of least privilege and separation of duties, by enabling Kerberos for Amazon EMR clusters."
  sql         = query.emr_cluster_kerberos_enabled.sql

  tags = merge(local.conformance_pack_emr_common_tags, {
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    nist_csf          = "true"
  })
}

control "emr_cluster_master_nodes_no_public_ip" {
  title       = "EMR cluster master nodes should not have public IP addresses"
  description = "Manage access to the AWS Cloud by ensuring Amazon EMR cluster master nodes cannot be publicly accessed."
  sql         = query.emr_cluster_master_nodes_no_public_ip.sql

  tags = merge(local.conformance_pack_emr_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}