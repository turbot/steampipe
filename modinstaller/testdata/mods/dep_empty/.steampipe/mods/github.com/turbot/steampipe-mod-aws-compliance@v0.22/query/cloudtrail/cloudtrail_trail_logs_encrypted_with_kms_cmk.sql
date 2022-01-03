select
  -- Required columns
  arn as resource,
  case
    when kms_key_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when kms_key_id is null then title || ' logs are not encrypted at rest.'
    else title || ' logs are encrypted at rest.'
  end as reason,
  -- Additional columns
  region,
  account_id
from
  aws_cloudtrail_trail
where 
  region = home_region;