select
  -- Required Columns
  arn as resource,
  case
    when deletion_protection then 'ok'
    else 'alarm'
  end status,
  case
    when deletion_protection then title || ' deletion protection enabled.'
    else title || ' deletion protection not enabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance;