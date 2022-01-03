select
  -- Required Columns
  arn as resource,
  case
    when alarm_actions is null
    and insufficient_data_actions is null
    and ok_actions is null then 'alarm'
    else 'ok'
  end as status,
  case
    when alarm_actions is null
    and insufficient_data_actions is null
    and ok_actions is null then title || ' no action enabled.'
    when alarm_actions is not null then title || ' alarm action enabled.'
    when insufficient_data_actions is not null then title || ' insufficient data action enabled.'
    when ok_actions is not null then title || ' ok action enabled.'
    else 'ok'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudwatch_alarm;
