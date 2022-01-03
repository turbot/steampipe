select
  -- Required Columns
  arn as resource,
  case
    when copy_tags_to_snapshot then 'ok'
    else 'alarm'
  end as status,
  case
    when copy_tags_to_snapshot then title || ' copy tags to snapshot enabled.'
    else title || ' copy tags to snapshot disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_cluster;
