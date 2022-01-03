select
  -- Required Columns
  arn as resource,
  case
    when versioning_mfa_delete then 'ok'
    else 'alarm'
  end status,
  case
    when versioning_mfa_delete then name || ' MFA delete enabled.'
    else name || ' MFA delete disabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_s3_bucket;