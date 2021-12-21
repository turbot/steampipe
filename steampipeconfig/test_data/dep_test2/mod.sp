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

  # dependencies
  require {
    steampipe = ">0.3.0"

    plugin "aws" {}

    plugin "gcp" {
      version = ">1.0.0"
    }

    # get by version tag
    mod  "github.com/turbot/aws-core" {
      version = "v1.123"
    }

    # get by tag and alias
    mod  "github.com/turbot/aws-core"  {
      version = "v2.345"
      alias = "aws_core_v2"
    }

    # get by branch
    mod  "github.com/turbot/aws-ec2-instance" {
      version = "staging"
    }

    # local mod
    mod "github.com/turbot/aws-ec2-elb" {
     version = "file:~/my_path/aws_core"
    }

  }
//  opengraph {
//    title = "Steampipe Mod for AWS CIS"
//    description = "CIS reports, queries, and actions for AWS. Open source CLI. No DB required."
//  }
}
