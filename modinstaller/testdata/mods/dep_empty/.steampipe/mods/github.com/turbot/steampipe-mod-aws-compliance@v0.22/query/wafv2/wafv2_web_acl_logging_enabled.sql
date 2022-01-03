select
  -- Required Columns
  arn as resource,
  case
    when logging_configuration is null then 'alarm'
    else 'ok'
  end as status,
  case
    when logging_configuration is null then title || ' logging disabled.'
    else title || ' logging enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_wafv2_web_acl;