mod "aws_compliance" {
  # hub metadata
  title          = "AWS Compliance"
  description    = "Steampipe Mod for Amazon Web Services (AWS) Compliance"
  color          = "#FF9900"
  documentation  = file("./docs/index.md")
  icon           = "/images/plugins/turbot/aws.svg"
  categories     = ["Public Cloud", "AWS"]

  opengraph {
    title         = "Steampipe Mod for AWS Compliance"
    description   = "Compliance and audit reports, queries, and actions for AWS. Open source CLI. No DB required."
  }

  /*
  # dependencies
  requires {
    steampipe ">0.3.0"
    plugin "aws"
  }
  */
}
