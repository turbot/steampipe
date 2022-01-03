benchmark "nist_800_53_rev_4_ir" {
  title       = "Incident Response (IR)"
  description = "IR controls are specific to an organization’s incident response policies and procedures. This includes incident response training, testing, monitoring, reporting, and response plan."
  children = [
    benchmark.nist_800_53_rev_4_ir_4,
    benchmark.nist_800_53_rev_4_ir_6,
    benchmark.nist_800_53_rev_4_ir_7
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ir_4" {
  title       = "Incident Handling (IR-4)"
  description = "The organization implements an incident handling capability for security incidents that includes preparation, detection and analysis, containment, eradication, and recovery, coordinates incident handling activities with contingency planning activities and incorporates lessons learned from ongoing incident handling activities into incident response procedures, training, and testing, and implements the resulting changes accordingly."
  children = [
    benchmark.nist_800_53_rev_4_ir_4_1
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ir_4_1" {
  title       = "IR-4(1) Automated Incident Handling Processes"
  description = "The organization employs automated mechanisms to support the incident handling process."
  children = [
    control.cloudwatch_alarm_action_enabled,
    control.guardduty_finding_archived
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ir_6" {
  title       = "Incident Reporting (IR-6)"
  description = "The organization report suspected security incidents to the organizational incident response capability within organization-defined time period."
  children = [
    benchmark.nist_800_53_rev_4_ir_6_1
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ir_6_1" {
  title       = "IR-6(1) Automated Reporting"
  description = "The organization employs automated mechanisms to assist in the reporting of security incidents."
  children = [
    control.guardduty_finding_archived
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ir_7" {
  title       = "Incident Response Assistance (IR-7)"
  description = "The organization provides an incident response support resource, integral to the organizational incident response capability that offers advice and assistance to users of the information system for the handling and reporting of security incidents."
  children = [
    benchmark.nist_800_53_rev_4_ir_7_1
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ir_7_1" {
  title       = "IR-7(1) Automation Support For Availability Of Information / Support"
  description = "The organization employs automated mechanisms to increase the availability of incident response-related information and support."
  children = [
    control.guardduty_finding_archived
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
