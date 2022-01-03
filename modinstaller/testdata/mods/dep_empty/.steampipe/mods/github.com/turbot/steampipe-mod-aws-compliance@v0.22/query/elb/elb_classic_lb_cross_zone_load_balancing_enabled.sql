select
  -- Required Columns
  arn as resource,
  case
    when cross_zone_load_balancing_enabled then 'ok'
    else 'alarm'
  end as status,
  case
    when cross_zone_load_balancing_enabled then title || ' cross-zone load balancing enabled.'
    else title || ' cross-zone load balancing disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_classic_load_balancer;