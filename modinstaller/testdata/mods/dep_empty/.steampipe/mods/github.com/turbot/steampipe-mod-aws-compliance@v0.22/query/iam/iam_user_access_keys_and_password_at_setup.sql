select
  -- Required Columns
  user_arn as resource,
  case
    -- alarm when password is enabled and the key was created within 10 seconds of the user
    when password_enabled and (extract(epoch from (access_key_1_last_rotated - user_creation_time)) < 10) then 'alarm'
    else 'ok'
  end as status,
  case
    when not password_enabled then user_name || ' password login disabled.'
    when access_key_1_last_rotated is null then user_name || ' has no access keys.'
    when password_enabled and (extract(epoch from (access_key_1_last_rotated - user_creation_time)) < 10)
      then user_name || ' has access key created during user creation and password login enabled.'
    else user_name || ' has access key not created during user creation.'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_credential_report
