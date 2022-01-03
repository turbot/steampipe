select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when logging_status ->> 'LoggingEnabled' = 'true' then 'ok'
    else 'alarm'
  end as status,
  case
    when logging_status ->> 'LoggingEnabled' = 'true' then title || ' logging enabled.'
    else title || ' logging disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;