dashboard "base_ref" {
  title = "With Graph as Node"

  input "instance_id" {
    title = "Select an instance:"
    query = query.ec2_instance_input
    width = 4
  }


  graph {


    node {
      base = node.ec2_instance
      args = {
        ec2_instance_ids = [self.input.instance_id.value]
      }
    }

  }
}


//************************

query "ec2_instance_input" {
  sql = <<-EOQ
    select
      title as label,
      instance_id as value,
      json_build_object(
        'account_id', account_id,
        'region', region,
        'instance_id', instance_id
      ) as tags
    from
      aws_ec2_instance
    order by
      title;
  EOQ
}
//************************




node "ec2_instance" {
  category = category.ec2_instance

  sql = <<-EOQ
    select
      instance_id as id,
      title,
      jsonb_build_object(
        'Instance ID', instance_id,
        'Name', tags ->> 'Name',
        'ARN', arn,
        'Account ID', account_id,
        'Region', region
      ) as properties
    from
      aws_ec2_instance
    where
      instance_id = any($1);
  EOQ

  param "ec2_instance_ids" {}
}



category "ec2_instance" {
  base = category.b
}

category "b" {
  title = "EC2 Instance"
  href  = "/aws_insights.dashboard.ec2_instance_detail?input.instance_arn={{.properties.'ARN' | @uri}}"
  icon  = "dns"

}
