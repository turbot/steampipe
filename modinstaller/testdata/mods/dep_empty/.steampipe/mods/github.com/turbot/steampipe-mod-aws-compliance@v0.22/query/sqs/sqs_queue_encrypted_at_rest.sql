select
  -- Required Columns
  queue_arn as resource,
  case
    when kms_master_key_id is null then 'alarm'
    else 'ok'
  end as status,
  case
    when kms_master_key_id is null then title || ' encryption at rest disabled.'
    else title || ' encryption at rest enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_sqs_queue;