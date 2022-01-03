select
  -- Required Columns
  arn as resource,
  case
    when date(last_accessed_date) - date(created_date) >= 1 then 'ok'
    else 'alarm'
  end as status,
  case
    when date(last_accessed_date)- date(created_date) >= 1 then title || ' recently used.'
    else title || ' not used recently.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_secretsmanager_secret;
