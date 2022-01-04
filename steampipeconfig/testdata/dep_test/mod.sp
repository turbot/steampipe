locals {
   l = ["aws", "cis", "cis_version:v1.3.0"]
   anythingelse = "doesnt work"
}

mod "m1" {
  # hub metadata
  title = local.anythingelse
  color = "#FF9900"
  icon = "/images/plugins/turbot/aws.svg"
  labels = local.l

  opengraph {
    title = "Steampipe Mod for AWS CIS"
    description = "CIS reports, queries, and actions for AWS. Open source CLI. No DB required."
  }
}
