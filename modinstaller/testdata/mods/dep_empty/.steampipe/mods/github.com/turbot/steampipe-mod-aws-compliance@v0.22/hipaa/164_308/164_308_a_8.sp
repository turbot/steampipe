benchmark "hipaa_164_308_a_8" {
  title       = "164.308(a)(8) Evaluation"
  description = "Perform a periodic technical and nontechnical evaluation, based initially upon the standards implemented under this rule and subsequently, in response to environmental or operational changes affecting the security of electronic protected health information, that establishes the extent to which an entity's security policies and procedures meet the requirements of this subpart."
  children = [
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = merge(local.hipaa_164_308_common_tags, {
    hipaa_item_id = "164_308_a_8"
  })
}
