select
  -- Required Columns
  user_arn as resource,
  case
    --root_account will have always password associated even though AWS credential report returns 'not_supported' for password_enabled
    when user_name = '<root_account>'
      then 'info'
    when password_enabled and password_last_used is null and password_last_changed < (current_date - interval '45' day)
      then 'alarm'
    when password_enabled and password_last_used  < (current_date - interval '45' day)
      then 'alarm'
    when access_key_1_active and access_key_1_last_used_date is null and access_key_1_last_rotated < (current_date - interval '45' day)
      then 'alarm'
    when access_key_1_active and access_key_1_last_used_date  < (current_date - interval '45' day)
      then 'alarm'
    when access_key_2_active and access_key_2_last_used_date is null and access_key_2_last_rotated < (current_date - interval '45' day)
      then 'alarm'
    when access_key_2_active and access_key_2_last_used_date  < (current_date - interval '45' day)
      then 'alarm'
    else 'ok'
  end status,
  user_name ||
    case
      when not password_enabled
        then ' password not enabled,'
      when password_enabled and password_last_used is null
        then ' password created ' || to_char(password_last_changed, 'DD-Mon-YYYY') || ' never used,'
      else
        ' password used ' || to_char(password_last_used, 'DD-Mon-YYYY') || ','
    end ||
    case
      when not access_key_1_active
        then ' key 1 not enabled,'
      when access_key_1_active and access_key_1_last_used_date is null
        then ' key 1 created ' || to_char(access_key_1_last_rotated, 'DD-Mon-YYYY') || ' never used,'
      else
        ' key 1 used ' || to_char(access_key_1_last_used_date, 'DD-Mon-YYYY') || ','
    end ||
      case
      when not access_key_2_active
        then ' key 2 not enabled.'
       when access_key_2_active and access_key_2_last_used_date is null
        then ' key 2 created ' || to_char(access_key_2_last_rotated, 'DD-Mon-YYYY') || ' never used.'
      else
        ' key 2 used ' || to_char(access_key_2_last_used_date, 'DD-Mon-YYYY') || '.'
    end
  as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_credential_report;