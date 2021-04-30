locals {
  l1 = "testing"
  q2 = "select 3 as Iamalocal"
}

mod "m1" {
  # hub metadata
  title = upper("aws cis")
  color = "#FF9900"
  icon = "/images/plugins/turbot/aws.svg"
  labels = ["public cloud", "aws"]

  opengraph {
    title = "Steampipe Mod for AWS CIS"
    description = "CIS reports, queries, and actions for AWS. Open source CLI. No DB required."
  }
}
