select
  -- Required Columns
  arn as resource,
  case
    when vpc_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when vpc_id is null then title || ' not in VPC.'
    else title || ' in VPC.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;