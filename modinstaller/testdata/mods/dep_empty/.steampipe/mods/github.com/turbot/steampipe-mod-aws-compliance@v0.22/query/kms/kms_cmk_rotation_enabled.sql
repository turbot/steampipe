select
  -- Required Columns
  arn as resource,
  case
    when origin = 'EXTERNAL' then 'skip'
    when not key_rotation_enabled then 'alarm'
    else 'ok'
  end as status,
  case
    when origin = 'EXTERNAL' then title || ' has imported key material.'
    when not key_rotation_enabled then title || ' key rotation disabled.'
    else title || ' key rotation enabled.'
  end as reason,
  -- Additional columns
  region,
  account_id
from
  aws_kms_key
where
  key_manager = 'CUSTOMER';
