select
  -- Required Columns
  user_arn as resource,
  case
    when not mfa_active then 'alarm'
    else 'ok'
  end as status,
  case
    when not mfa_active then user_name || ' MFA device not configured.'
    else user_name || ' MFA device configured.'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_credential_report;