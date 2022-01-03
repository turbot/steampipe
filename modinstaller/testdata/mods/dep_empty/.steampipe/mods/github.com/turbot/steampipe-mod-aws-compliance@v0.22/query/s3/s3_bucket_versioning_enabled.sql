select
  -- Required Columns
  arn as resource,
  case
    when versioning_enabled then 'ok'
    else 'alarm'
  end status,
  case
    when versioning_enabled then name || ' versioning enabled.'
    else name || ' versioning disabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_s3_bucket