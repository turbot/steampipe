select
  -- Required Columns
  arn as resource,
  case
    when backup_retention_period < 1 then 'alarm'
    else 'ok'
  end as status,
  case
    when backup_retention_period < 1 then title || ' backups not enabled.'
    else title || ' backups enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance;