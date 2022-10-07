dashboard "aws_ec2_instance_detail" {
  title         = "AWS EC2 Instance Detail"

    graph {
      type      = "graph"
      direction = "LR"

      # no base, resource named
      category "c1" {
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
      }

    }
    graph {
      type      = "graph"
      direction = "LR"

      # no base, resource named
      category "c1" {
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
      }
      # no base, named
      category {
        name = "aws_ebs_volume2"
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
      }
      #  no name - will fail
#      category {
#        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
#      }
      # base no name - will use base name
      category {
        base = category.aws_ebs_volume
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
        fold {
          title     = "EC2 Classic Load Balancers"
          threshold = 3
        }
      }
      # base with name
      category {
        base = category.aws_ebs_volume
        name = "aws_ebs_volume2"
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
        fold {
          title     = "EC2 Classic Load Balancers"
          threshold = 3
        }
      }
    }
}

# has name - ignores resource name for
category "aws_ec2_instance" {

  href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
  fold {
    title     = "EC2 Classic Load Balancers"
    threshold = 3
  }
}
# has name - ignores resource name for
category "top_level_category_with_explicit_name" {
  name = "aws_s3_bucket"
  href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
  fold {
    title     = "EC2 Classic Load Balancers"
    threshold = 3
  }
}

# no name - will use resource short
category "aws_ebs_volume" {
  href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
  fold {
    title     = "EC2 Classic Load Balancers"
    threshold = 3
  }
}