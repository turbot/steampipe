select
  -- Required Columns
  arn as resource,
  case
    when key_state = 'PendingDeletion' then 'alarm'
    else 'ok'
  end as status,
  case
    when key_state = 'PendingDeletion' then title || ' scheduled for deletion and will be deleted in ' || extract(day from deletion_date - current_timestamp) || ' day(s).'
    else title || ' not scheduled for deletion.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_kms_key
where
  key_manager = 'CUSTOMER';