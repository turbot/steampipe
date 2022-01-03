select
  -- Required Columns
  arn as resource,
  case
    when engine ilike any (array ['%aurora-mysql%', '%aurora-postgres%']) then 'skip'
    when multi_az then 'ok'
    else 'alarm'
  end as status,
  case
    when engine ilike any (array ['%aurora-mysql%', '%aurora-postgres%']) then title || ' cluster instance.'
    when multi_az then title || ' Multi-AZ enabled.'
    else title || ' Multi-AZ disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance;
