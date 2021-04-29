mod "m1" {
  # hub metadata
  title = upper("aws cis")
  description = control.c1.description
  color = "#FF9900"
  icon = "/images/plugins/turbot/aws.svg"
  labels = [
    "public cloud",
    "aws"]

  opengraph {
    title = "Steampipe Mod for AWS CIS"
    description = "CIS reports, queries, and actions for AWS. Open source CLI. No DB required."

  }
}