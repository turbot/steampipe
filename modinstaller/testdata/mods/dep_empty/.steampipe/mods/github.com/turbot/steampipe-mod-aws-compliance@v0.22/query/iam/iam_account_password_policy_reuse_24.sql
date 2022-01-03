select
  -- Required Columns
  'arn:' || a.partition || ':::' || a.account_id as resource,
  case
    when password_reuse_prevention >= 24 then 'ok'
    else 'alarm'
  end as status,
  case
    when minimum_password_length is null then 'No password policy set.'
    when password_reuse_prevention is null then 'Password reuse prevention not set.'
    else 'Password reuse prevention set to ' || password_reuse_prevention || '.'
  end as reason,
  -- Additional Dimensions
  a.account_id
from
  aws_account as a
  left join aws_iam_account_password_policy as pol on a.account_id = pol.account_id;
