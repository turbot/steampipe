select
  -- Required Columns
  arn as resource,
  case
    when reserved_concurrent_executions is null then 'alarm'
    else 'ok'
  end as status,
  case
    when reserved_concurrent_executions is null then title || ' function-level concurrent execution limit not configured.'
    else title || ' function-level concurrent execution limit configured.'
  end as reason,
  -- Additional Columns
  region,
  account_id
from
  aws_lambda_function;