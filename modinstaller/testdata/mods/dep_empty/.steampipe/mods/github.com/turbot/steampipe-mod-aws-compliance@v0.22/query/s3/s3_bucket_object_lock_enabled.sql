select
  -- Required Columns
  arn as resource,
  case
    when object_lock_configuration is null then 'alarm'
    else 'ok'
  end as status,
  case
    when object_lock_configuration is null then title || ' object lock not enabled.'
    else title || ' object lock enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_s3_bucket;