benchmark "rbi_cyber_security_annex_i_6" {
  title       = "Annex I (6)"
  description = "Put in place systems and processes to identify, track, manage and monitor the status of patches to servers, operating system and application software running at the systems used by the UCB officials (end-users). Implement and update antivirus protection for all servers and applicable end points preferably through a centralised system."

  children = [
    control.guardduty_finding_archived,
    control.rds_db_instance_automatic_minor_version_upgrade_enabled,
    control.redshift_cluster_maintenance_settings_check,
    control.ssm_managed_instance_compliance_association_compliant,
    control.ssm_managed_instance_compliance_patch_compliant
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_6"
  })
}
