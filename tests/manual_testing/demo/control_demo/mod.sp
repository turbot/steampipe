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
//
//  require {
//    steampipe  ">0.3.0" {}
//
//    plugin "aws" {}
//    plugin "gcp" ">1.0.0" {}
//
//    # get by version tag
//    mod  "github.com/turbot/aws-core" "v1.123" {}
//
//    # get by tag and alias
//    mod  "github.com/turbot/aws-core" "v2.345" {
//      alias = "aws_core_v2"
//    }
//
//    # get by branch
//    mod  "github.com/turbot/aws-ec2-instance" "staging" {}
//
//    # local mod
//    mod "github.com/turbot/aws-ec2-elb" "file:~/my_path/aws_core"{}
//
//  }
}