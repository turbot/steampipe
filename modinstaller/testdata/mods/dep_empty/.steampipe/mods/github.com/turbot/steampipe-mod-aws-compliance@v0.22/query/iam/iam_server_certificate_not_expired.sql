select
  -- Required Columns
  arn as resource,
  case when expiration < (current_date - interval '1' second) then 'alarm'
  else 'ok'
  end as status,
  case when expiration < (current_date - interval '1' second) then
    name || ' expired ' || to_char(expiration, 'DD-Mon-YYYY') || '.'
  else
    name || ' valid until ' || to_char(expiration, 'DD-Mon-YYYY')  || '.'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_server_certificate;