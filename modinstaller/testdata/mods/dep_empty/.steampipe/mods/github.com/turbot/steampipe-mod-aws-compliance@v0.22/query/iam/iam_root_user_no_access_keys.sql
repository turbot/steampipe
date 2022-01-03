select
  -- Required Columns
  'arn:' || partition || ':::' || account_id as resource,
  case
    when account_access_keys_present > 0 then 'alarm'
    else 'ok'
  end status,
  case
    when account_access_keys_present > 0 then 'Root user access keys exist.'
    else 'No root user access keys exist.'
  end reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_account_summary;
