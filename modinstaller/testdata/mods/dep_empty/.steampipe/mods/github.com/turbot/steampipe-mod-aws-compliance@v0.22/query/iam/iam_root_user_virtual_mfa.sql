select
  -- Required Columns
  'arn:' || s.partition || ':::' || s.account_id as resource,
  case
    when account_mfa_enabled and serial_number is not null then 'ok'
    else 'alarm'
  end status,
  case
    when account_mfa_enabled = false then 'MFA is not enabled for the root user.'
    when serial_number is null then 'MFA is enabled for the root user, but the MFA associated with the root user is a hardware device.'
    else 'Virtual MFA enabled for the root user.'
  end reason,
  -- Additional Dimensions
  s.account_id
from
  aws_iam_account_summary as s
  left join aws_iam_virtual_mfa_device on serial_number = 'arn:' || s.partition || ':iam::' || s.account_id || ':mfa/root-account-mfa-device';