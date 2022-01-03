select
  -- Required Columns
  arn as resource,
  case
    when server_side_encryption_configuration is not null then 'ok'
    else 'alarm'
  end status,
  case
    when server_side_encryption_configuration is not null then name || ' default encryption enabled.'
    else name || ' default encryption disabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_s3_bucket