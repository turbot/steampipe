

node "aws_ec2_instance_node" {

  with "w1" {
    sql = "select 1"
    param "instance_id" {}
  }

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
