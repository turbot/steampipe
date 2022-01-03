select
  -- Required Columns
  'arn:' || a.partition || ':::' || a.account_id as resource,
  case
    when require_symbols then 'ok'
    else 'alarm'
  end as status,
  case
    when minimum_password_length is null then 'No password policy set.'
    when require_symbols then 'Symbol required.'
    else 'Symbol not required.'
  end as reason,
  -- Additional Dimensions
  a.account_id
from
  aws_account as a
  left join aws_iam_account_password_policy as pol on a.account_id = pol.account_id;