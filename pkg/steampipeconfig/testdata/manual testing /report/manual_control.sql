select
  -- Required Columns
  'arn:' || partition || ':::' || account_id as resource,
  'info' as status,
  'Manual verification required.' as reason,
  -- Additional Dimensions
  account_id
from
  aws_account;
