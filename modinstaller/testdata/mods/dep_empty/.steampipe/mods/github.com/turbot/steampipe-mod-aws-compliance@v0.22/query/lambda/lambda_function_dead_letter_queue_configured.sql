select
  -- Required Columns
  arn as resource,
  case
    when dead_letter_config_target_arn is null then 'alarm'
    else 'ok'
  end as status,
  case
    when dead_letter_config_target_arn is null then title || ' configured with dead-letter queue.'
    else title || ' not configured with dead-letter queue.'
  end as reason,
  -- Additional Columns
  region,
  account_id
from
  aws_lambda_function;