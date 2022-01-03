select
  -- Required Columns
  arn as resource,
  case
    when rotation_rules is null then 'alarm'
    else 'ok'
  end as status,
  case
    when rotation_rules is null then title || ' automatic rotation not enabled.'
    else title || ' automatic rotation enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_secretsmanager_secret;
