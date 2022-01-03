select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when automated_snapshot_retention_period >= 7 then 'ok'
    else 'alarm'
  end as status,
  case
    when automated_snapshot_retention_period >= 7 then title || ' automatic snapshots enabled with retention period greater than equals 7 days.'
    else title || ' automatic snapshots not enabled with retention period greater than equals 7 days.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;