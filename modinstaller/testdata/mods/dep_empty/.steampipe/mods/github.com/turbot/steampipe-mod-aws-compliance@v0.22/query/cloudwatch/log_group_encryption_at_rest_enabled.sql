select
  -- Required Columns
  arn as resource,
  case
    when kms_key_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when kms_key_id is null then title || ' not encrypted at rest.'
    else title || ' encrypted at rest.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudwatch_log_group;