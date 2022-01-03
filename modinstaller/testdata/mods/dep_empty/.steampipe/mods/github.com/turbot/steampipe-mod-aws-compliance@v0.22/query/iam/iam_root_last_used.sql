select
  -- Required Columns
  user_arn as resource,
  case
    when password_last_used >= (current_date - interval '90' day) then 'alarm'
    when access_key_1_last_used_date <= (current_date - interval '90' day)  then 'alarm'
    when access_key_2_last_used_date <= (current_date - interval '90' day)  then 'alarm'
    else 'ok'
  end as status,
  case
    when password_last_used is null then 'Root never logged in with password.'
    else 'Root password used ' || to_char(password_last_used , 'DD-Mon-YYYY') || ' (' || extract(day from current_timestamp - password_last_used) || ' days).'
  end ||
  case
    when access_key_1_last_used_date is null then ' Access Key 1 never used.'
    else ' Access Key 1 used ' || to_char(access_key_1_last_used_date , 'DD-Mon-YYYY') || ' (' || extract(day from current_timestamp - access_key_1_last_used_date) || ' days).'
  end ||
    case
    when access_key_2_last_used_date is null then ' Access Key 2 never used.'
    else ' Access Key 2 used ' || to_char(access_key_2_last_used_date , 'DD-Mon-YYYY') || ' (' || extract(day from current_timestamp - access_key_2_last_used_date) || ' days).'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_credential_report
where
  user_name = '<root_account>';