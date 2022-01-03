select
  -- Required Columns
  'arn:' || partition || ':iam::' || account_id || ':user/' || user_name || '/accesskey/' || access_key_id as resource,
  case
    when create_date <= (current_date - interval '90' day) then 'alarm'
    else 'ok'
  end status,
  user_name || ' ' || access_key_id || ' created ' || to_char(create_date , 'DD-Mon-YYYY') ||
    ' (' || extract(day from current_timestamp - create_date) || ' days).'
  as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_access_key;
