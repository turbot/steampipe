select
  -- Required Columns
  'arn:' || a.partition || ':::' || a.account_id as resource,
  case
    when
      minimum_password_length >= 8
      and require_lowercase_characters = 'true'
      and require_uppercase_characters = 'true'
      and require_numbers = 'true'
      and require_symbols = 'true'
    then 'ok'
    else 'alarm'
  end as status,
  case
    when minimum_password_length is null then 'No password policy set.'
    when
      minimum_password_length >= 8
      and require_lowercase_characters = 'true'
      and require_uppercase_characters = 'true'
      and require_numbers = 'true'
      and require_symbols = 'true'
    then 'Strong password policies configured.'
    else 'Strong password policies not configured.'
  end as reason,
  -- Additional Dimensions
  a.account_id
from
  aws_account as a
  left join aws_iam_account_password_policy as pol on a.account_id = pol.account_id;
