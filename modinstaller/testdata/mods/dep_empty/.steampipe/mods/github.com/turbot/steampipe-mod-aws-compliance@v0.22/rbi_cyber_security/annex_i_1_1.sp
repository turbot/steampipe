benchmark "rbi_cyber_security_annex_i_1_1" {
  title       = "Annex I (1.1)"
  description = "UCBs should maintain an up-to-date business IT Asset Inventory Register containing the following fields, as a minimum: a) Details of the IT Asset (viz., hardware/software/network devices, key personnel, services, etc.), b. Details of systems where customer data are stored, c. Associated business applications, if any, d. Criticality of the IT asset (For example, High/Medium/Low)."

  children = [
    control.ec2_instance_ssm_managed
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_1_1"
  })
}
