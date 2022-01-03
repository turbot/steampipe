select
  -- Required Columns
  arn as resource,
  case
    when vpc_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when vpc_id is null then title || ' is not in VPC.'
    else title || ' is in VPC ' || vpc_id || '.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance;