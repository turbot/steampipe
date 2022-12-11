dashboard "with_graph_as_node" {
  title = "With Graph as Node"

  input "instance_id" {
    title = "Select an instance:"
    query = query.ec2_instance_input
    width = 4
  }

  //self.input.instance_id.value
  //i-04cd720cb5b0e2d05

  with "security_groups" {
    sql = <<-EOQ
      select
        s ->> 'GroupId' as sg_id
      from
        aws_ec2_instance,
        jsonb_array_elements(security_groups) as s
      where
        instance_id = $1
    EOQ

    args = [self.input.instance_id.value]
  }

  with "vpc_details" {
    sql = <<-EOQ
      select
        instance_id,
        vpc_id,
        subnet_id
      from
        aws_ec2_instance
      where
        instance_id = $1
    EOQ

    args = [self.input.instance_id.value]
  }


  graph {


    node {
      base = node.ec2_instance
      args = {
        ec2_instance_ids  = [self.input.instance_id.value]
      }
    }

    edge {
      base = edge.aws_ec2_instance_to_security_group
      args = {
        ec2_instance_id  = self.input.instance_id.value
      }
    }



    node {
      base = node.vpc_vpc
      args = {
        vpc_vpc_ids = with.vpc_details.rows[*].vpc_id
      }
    }

    node "mynode" {
      base = node.vpc_subnet
      args = {
        vpc_subnet_ids    = with.vpc_details.rows[*].subnet_id
      }
    }

    node {
      base = node.vpc_security_group
      args = {
        vpc_security_group_ids        = with.security_groups.rows[*].sg_id
      }
    }

    edge {
      base = edge.vpc_security_group_to_vpc_subnet
      args = {
        vpc_security_group_ids        = with.security_groups.rows[*].sg_id
      }
    }

    edge {
      base = edge.vpc_subnet_to_vpc
      args = {
        vpc_subnet_ids    = with.vpc_details.rows[*].subnet_id
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
  //category = category.ec2_instance

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



edge "aws_ec2_instance_to_security_group" {
  title = "security group"

  sql = <<-EOQ
    select
      instance_id as from_id,
      sg ->> 'GroupId' as to_id
    from
      aws_ec2_instance,
      jsonb_array_elements(security_groups) as sg
    where
      instance_id = $1
  EOQ

  param "ec2_instance_id" {}
}

//**********************






node "vpc_vpc" {
  //category = category.vpc_vpc

  sql = <<-EOQ
   select
      vpc_id as id,
      title as title,
      jsonb_build_object(
        'ARN', arn,
        'VPC ID', vpc_id,
        'Is Default', is_default,
        'State', state,
        'CIDR Block', cidr_block,
        'DHCP Options ID', dhcp_options_id,
        'Owner ID', owner_id,
        'Account ID', account_id,
        'Region', region
      ) as properties
    from
      aws_vpc
    where
      vpc_id = any($1 ::text[]);
  EOQ

  param "vpc_vpc_ids" {}
}



node "vpc_security_group" {
  //category = category.vpc_security_group

  sql = <<-EOQ
    select
      group_id as id,
      title as title,
      jsonb_build_object(
        'Group ID', group_id,
        'Description', description,
        'ARN', arn,
        'Account ID', account_id,
        'Region', region
      ) as properties
    from
      aws_vpc_security_group
    where
      group_id = any($1 ::text[]);
  EOQ

  param "vpc_security_group_ids" {}
}

node "vpc_subnet" {
  //category = category.vpc_subnet

  sql = <<-EOQ
   select
      subnet_id as id,
      title as title,
      jsonb_build_object(
        'Subnet ID', subnet_id,
        'ARN', subnet_arn,
        'VPC ID', vpc_id,
        'Account ID', account_id,
        'Region', region
      ) as properties
    from
      aws_vpc_subnet
    where
      subnet_id = any($1 ::text[]);
  EOQ

  param "vpc_subnet_ids" {}
}




////////
edge "vpc_subnet_to_vpc" {
  title = "vpc"

  sql = <<-EOQ
    select
      subnet_id as from_id,
      vpc_id as to_id
    from
      aws_vpc_subnet
    where
      subnet_id = any($1)
  EOQ

  param "vpc_subnet_ids" {}
}


//edge.aws_security_group_to_subnets

edge "vpc_security_group_to_vpc_subnet" {
  title = "subnet"

  sql = <<-EOQ
    select
      subnet.subnet_id as from_id,
      sg.group_id as to_id
    from
      aws_vpc_security_group as sg,
      aws_svpc_subnet as subnet
    where
      sg.vpc_id = subnet.vpc_id
      and sg.group_id = any($1)
  EOQ

  param "vpc_security_group_ids" {}
}