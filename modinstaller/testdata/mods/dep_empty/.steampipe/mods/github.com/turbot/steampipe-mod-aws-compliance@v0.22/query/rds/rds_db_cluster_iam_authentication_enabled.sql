select
  -- Required Columns
  arn as resource,
  case
    when iam_database_authentication_enabled then 'ok'
    else 'alarm'
  end as status,
  case
    when iam_database_authentication_enabled then title || ' IAM authentication enabled.'
    else title || ' IAM authentication not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_cluster;