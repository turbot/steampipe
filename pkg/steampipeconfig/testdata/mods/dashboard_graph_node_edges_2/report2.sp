dashboard "aws_ec2_instance_detail" {
  title = "AWS EC2 Instance Detail"

  graph {
    type      = "graph"
    direction = "LR"

    # no base
    category "c1" {
      href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
    }
    # base
    category "aws_ebs_volume" {
      base = category.aws_ebs_volume
      href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
      fold {
        title     = "EC2 Classic Load Balancers"
        threshold = 3
      }
    }

  }
}

category "aws_ec2_instance" {

  href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
  fold {
    title     = "EC2 Classic Load Balancers"
    threshold = 3
  }
}

category "aws_ebs_volume" {
  href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
  fold {
    title     = "EC2 Classic Load Balancers"
    threshold = 3
  }
}
