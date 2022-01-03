select
  -- Required Columns
  user_arn as resource,
  case
    when password_enabled and not mfa_active then 'alarm'
    else 'ok'
  end as status,
  case
    when not password_enabled then user_name || ' password login disabled.'
    when password_enabled and not mfa_active then user_name || ' password login enabled but no MFA device configured.'
    else user_name || ' password login enabled and MFA device configured.'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_credential_report;