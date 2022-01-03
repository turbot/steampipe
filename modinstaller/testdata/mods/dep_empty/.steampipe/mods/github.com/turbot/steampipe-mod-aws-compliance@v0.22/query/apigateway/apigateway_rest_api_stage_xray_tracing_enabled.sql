select
  -- Required Columns
  arn as resource,
  case
    when tracing_enabled then 'ok'
    else 'alarm'
  end as status,
  case
    when tracing_enabled then title || ' X-Ray tracing enabled.'
    else title || ' X-Ray tracing disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_api_gateway_stage;