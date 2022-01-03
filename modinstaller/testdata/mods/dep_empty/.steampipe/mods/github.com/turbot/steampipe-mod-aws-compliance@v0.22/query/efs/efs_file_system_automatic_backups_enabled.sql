select
  -- Required Columns
  arn as resource,
  case
    when automatic_backups = 'enabled' then 'ok'
    else 'alarm'
  end as status,
  case
    when automatic_backups = 'enabled' then title || ' automatic backups enabled.'
    else title || ' automatic backups not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_efs_file_system;
