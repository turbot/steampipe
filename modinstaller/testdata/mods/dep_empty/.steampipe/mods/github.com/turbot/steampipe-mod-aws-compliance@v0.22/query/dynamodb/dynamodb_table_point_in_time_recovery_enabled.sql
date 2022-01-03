select
  -- Required Columns
  arn as resource,
  case
    when lower( point_in_time_recovery_description ->> 'PointInTimeRecoveryStatus' ) = 'disabled' then 'alarm'
    else 'ok'
  end as status,
  case
    when lower( point_in_time_recovery_description ->> 'PointInTimeRecoveryStatus' ) = 'disabled' then title || ' point-in-time recovery not enabled.'
    else title || ' point-in-time recovery enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_dynamodb_table;
