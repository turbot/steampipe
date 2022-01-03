select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when enhanced_vpc_routing then 'ok'
    else 'alarm'
  end as status,
  case
    when enhanced_vpc_routing then title || ' enhanced VPC routing enabled.'
    else title || ' enhanced VPC routing disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster;