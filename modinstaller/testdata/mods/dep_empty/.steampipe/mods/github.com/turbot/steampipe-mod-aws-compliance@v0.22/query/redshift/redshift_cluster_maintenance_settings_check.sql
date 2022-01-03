select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when allow_version_upgrade and automated_snapshot_retention_period >= 7 then 'ok'
    else 'alarm'
  end as status,
  case
    when allow_version_upgrade and automated_snapshot_retention_period >= 7 then title || ' has the required maintenance settings.'
    else title || ' does not have required maintenance settings.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;