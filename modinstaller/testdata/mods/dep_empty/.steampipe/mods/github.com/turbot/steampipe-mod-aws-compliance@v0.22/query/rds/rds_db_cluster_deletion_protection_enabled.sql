 select
  -- Required Columns
  db_cluster_identifier as resource,
  case
    when deletion_protection then 'ok'
    else 'alarm'
  end as status,
  case
    when deletion_protection then title || ' deletion protection enabled.'
    else title || ' deletion protection not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_cluster;