select
  -- Required Columns
  arn as resource,
  case
    when log_file_validation_enabled then 'ok'
    else 'alarm'
  end as status,
  case
    when log_file_validation_enabled then title || ' log file validation enabled.'
    else title || ' log file validation disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudtrail_trail
where 
  region = home_region;