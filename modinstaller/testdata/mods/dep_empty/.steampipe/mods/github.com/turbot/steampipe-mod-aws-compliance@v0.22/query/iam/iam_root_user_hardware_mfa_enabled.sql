select
  -- Required Columns
  'arn:' || s.partition || ':::' || s.account_id as resource,
  case
    when account_mfa_enabled and serial_number is null then 'ok'
    else 'alarm'
  end status,
  case
    when account_mfa_enabled = false then  'MFA not enabled for root account.'
    when serial_number is not null then 'Virtual MFA device enabled the root account.'
    else 'Hardware MFA device enabled for root account.'
  end reason,
  -- Additional Dimensions
  s.account_id
from
  aws_iam_account_summary as s
  left join aws_iam_virtual_mfa_device on serial_number = 'arn:' || s.partition || ':iam::' || s.account_id || ':mfa/root-account-mfa-device'