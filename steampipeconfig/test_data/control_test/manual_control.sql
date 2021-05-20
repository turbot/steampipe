select
  -- required columns
  'arn:' || partition || ':::' || account_id as resource,
  'info' as status,
  'This is a manual control, you must verify compliance manually.' as reason,
  -- extra columns (annotations?)
  account_id,
  partition,
  region
from
  aws_account;
