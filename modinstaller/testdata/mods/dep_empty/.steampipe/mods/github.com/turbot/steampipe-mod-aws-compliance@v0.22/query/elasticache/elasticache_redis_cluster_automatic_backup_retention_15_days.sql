select
  -- Required Columns
  arn as resource,
  case
    when snapshot_retention_limit < 15 then 'alarm'
    else 'ok'
  end as status,
  case
    when snapshot_retention_limit = 0 then title || ' automatic backups not enabled.'
    when snapshot_retention_limit < 15 then title || ' automatic backup retention period is less than 15 days.'
    else title || ' automatic backup retention period is more than 15 days.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticache_replication_group;
