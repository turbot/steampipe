select
  -- Required Columns
  arn as resource,
  case
    when encryption_at_rest_options ->> 'Enabled' = 'false' then 'alarm'
    else 'ok'
  end status,
  case
    when encryption_at_rest_options ->> 'Enabled' = 'false' then title || ' encryption at rest not enabled.'
    else title || ' encryption at rest enabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;
