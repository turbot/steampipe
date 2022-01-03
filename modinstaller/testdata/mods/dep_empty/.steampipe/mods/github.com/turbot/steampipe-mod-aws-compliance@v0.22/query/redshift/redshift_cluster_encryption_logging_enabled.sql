select
  -- Required Columns
  arn as resource,
  case
    when not encrypted then 'alarm'
    when not (logging_status ->> 'LoggingEnabled') :: boolean then 'alarm'
    else 'ok'
  end as status,
  case
    when not encrypted then title || ' not encrypted.'
    when not (logging_status ->> 'LoggingEnabled') :: boolean then title || ' audit logging not enabled.'
    else title || ' audit logging and encryption enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;