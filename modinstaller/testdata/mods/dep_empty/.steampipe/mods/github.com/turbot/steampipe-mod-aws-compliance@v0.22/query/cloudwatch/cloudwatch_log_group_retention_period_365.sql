select
  -- Required Columns
  arn as resource,
  case
    when retention_in_days is null or retention_in_days < 365 then 'alarm'
    else 'ok'
  end as status,
  case
    when retention_in_days is null then title || ' retention period not set.'
    when retention_in_days < 365 then title || ' retention period less than 365 days.'
    else title || ' retention period 365 days or above.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudwatch_log_group;