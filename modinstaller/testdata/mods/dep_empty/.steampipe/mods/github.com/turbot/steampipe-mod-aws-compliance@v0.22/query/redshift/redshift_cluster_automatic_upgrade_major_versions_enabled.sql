select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when allow_version_upgrade then 'ok'
    else 'alarm'
  end as status,
  case
    when allow_version_upgrade then title || ' automatic upgrades to major versions enabled.'
    else title || ' automatic upgrades to major versions disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;