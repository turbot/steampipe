benchmark "nist_800_53_rev_4_ra" {
  title       = "Risk Assessment (RA)"
  description = "The RA control family relates to an organization’s risk assessment policies and vulnerability scanning capabilities. Using an integrated risk management solution like CyberStrong can help streamline and automate your NIST 800 53 compliance efforts."
  children = [
    benchmark.nist_800_53_rev_4_ra_5
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ra_5" {
  title       = "Vulnerability Scanning (RA-5)"
  description = "Scan for system vulnerabilities. Share vulnerability information and security controls that eliminate vulnerabilities."
  children = [
    control.guardduty_enabled,
    control.guardduty_finding_archived
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
