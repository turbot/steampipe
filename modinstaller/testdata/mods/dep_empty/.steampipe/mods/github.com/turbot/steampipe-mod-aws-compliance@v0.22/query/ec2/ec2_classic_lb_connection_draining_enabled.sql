select
  -- Required Columns
  arn as resource,
  case
    when connection_draining_enabled then 'ok'
    else 'alarm'
  end as status,
  case
    when connection_draining_enabled then title || ' connection draining enabled.'
    else title || ' connection draining disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_classic_load_balancer;