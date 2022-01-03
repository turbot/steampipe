select
  -- Required Columns
  arn as resource,
  case
    when logging ->> 'Enabled' = 'true' then 'ok'
    else 'alarm'
  end as status,
  case
    when logging ->> 'Enabled' = 'true' then title || ' logging enabled.'
    else title || ' logging disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudfront_distribution;