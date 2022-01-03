select
  -- Required Columns
  arn as resource,
  case
    when log_group_arn != 'null' and ((latest_delivery_time) > current_date - 1) then 'ok'
    else 'alarm'
  end as status,
  case
    when log_group_arn != 'null' and ((latest_delivery_time) > current_date - 1) then title || ' integrated with CloudWatch logs.'
    else title || ' not integrated with CloudWatch logs.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudtrail_trail
where 
  region = home_region;