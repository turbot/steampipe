dashboard "aws_ec2_instance_detail" {
  title         = "AWS EC2 Instance Detail"

  input "instance_id" {
    title = "Select an instance:"
    sql   = query.aws_ec2_instance_input.sql
    width = 4
  }

  container {

    graph {
      type  = "graph"
      direction = "LR"

      nodes = [
        node.aws_ec2_instance_node,
        node.aws_ec2_instance_ebs_volume_node
      ]

      edges = [
        edge.aws_ec2_instance_to_ebs_volume_edge
      ]

      category  {
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
        fold {
          title     = "EC2 Classic Load Balancers"
          threshold = 3
        }
      }
      category  {
        href = "/aws_insights.dashboard.aws_ec2_classic_load_balancer_detail?input.clb={{.properties.'ARN' | @uri}}"
        fold {
          title     = "EC2 Classic Load Balancers"
          threshold = 3
        }
      }


      param "instance_id" {}
      args = {
        instance_id = self.input.instance_id.value
      }
    }
  }

}

query "aws_ec2_instance_input" {
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


node "aws_ec2_instance_node" {
  category = category.aws_ec2_instance

  sql = <<-EOQ
    select
      instance_id as id,
      title as title,
      jsonb_build_object(
        'Name', tags ->> 'Name',
        'Instance ID', instance_id,
        'ARN', arn,
        'Account ID', account_id,
        'Region', region
      ) as properties
    from
        aws_ec2_instance
      where
        instance_id = $1
  EOQ

  param "instance_id" {}
}



node "aws_ec2_instance_ebs_volume_node" {
  category = category.aws_ebs_volume

  sql = <<-EOQ
    select
      bd -> 'Ebs' ->> 'VolumeId' as id,
      bd -> 'Ebs' ->> 'VolumeId' as title,
      jsonb_build_object(
        'ARN', v.arn,
        'Account ID', v.account_id,
        'Region', v.region,
        'Volume ID', bd -> 'Ebs' ->> 'VolumeId'
      ) as properties
    from
      aws_ec2_instance as i,
      jsonb_array_elements(block_device_mappings) as bd,
      aws_ebs_volume as v
    where
      i.instance_id = $1
      and v.volume_id = bd -> 'Ebs' ->> 'VolumeId'
  EOQ

  param "instance_id" {}
}


edge "aws_ec2_instance_to_ebs_volume_edge" {
  title = "mounts"

  sql = <<-EOQ
    select
      instance_id as from_id,
      bd -> 'Ebs' ->> 'VolumeId' as to_id,
      jsonb_build_object(
        'Account ID', i.account_id,
        'Volume ID', bd -> 'Ebs' ->> 'VolumeId',
        'Device Name', bd ->> 'DeviceName',
        'Status', bd -> 'Ebs' ->> 'Status',
        'Attach Time', bd -> 'Ebs' ->> 'AttachTime',
        'Delete On Termination', bd -> 'Ebs' ->> 'DeleteOnTermination'
      ) as properties
    from
      aws_ec2_instance as i,
      jsonb_array_elements(block_device_mappings) as bd
    where
      i.instance_id = $1
  EOQ

  param "instance_id" {}
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