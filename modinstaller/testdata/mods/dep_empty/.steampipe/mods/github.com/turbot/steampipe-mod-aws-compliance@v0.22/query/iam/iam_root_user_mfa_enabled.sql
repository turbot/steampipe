select
  -- Required Columns
  'arn:' || partition || ':::' || account_id as resource,
  case
    when account_mfa_enabled then 'ok'
    else 'alarm'
  end status,
  case
    when account_mfa_enabled then 'MFA enabled for root account.'
    else 'MFA not enabled for root account.'
  end reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_account_summary;
