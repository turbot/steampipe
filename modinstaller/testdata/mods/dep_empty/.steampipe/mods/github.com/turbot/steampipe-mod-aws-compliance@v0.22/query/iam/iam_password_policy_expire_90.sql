select
  -- Required Columns
  'arn:' || a.partition || ':::' || a.account_id as resource,
  case
    when max_password_age <= 90 then 'ok'
    else 'alarm'
  end as status,
  case
    when max_password_age is null then 'Password expiration not set.'
    else 'Password expiration set to ' || max_password_age || ' days.'
  end as reason,
  -- Additional Dimensions
  a.account_id
from
  aws_account as a
  left join aws_iam_account_password_policy as pol on a.account_id = pol.account_id;