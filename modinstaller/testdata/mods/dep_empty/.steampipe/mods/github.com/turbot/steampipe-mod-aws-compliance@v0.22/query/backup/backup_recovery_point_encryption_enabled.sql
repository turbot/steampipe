select
  -- Required Columns
  recovery_point_arn as resource,
  case
    when is_encrypted then 'ok'
    else 'alarm'
  end as status,
  case
    when is_encrypted then recovery_point_arn || ' encryption enabled.'
    else recovery_point_arn || ' encryption disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_backup_recovery_point;