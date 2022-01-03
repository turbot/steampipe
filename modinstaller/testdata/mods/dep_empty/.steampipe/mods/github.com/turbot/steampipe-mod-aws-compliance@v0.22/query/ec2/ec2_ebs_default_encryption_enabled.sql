select
  -- Required Columns
  'arn:' || partition || '::' || region || ':' || account_id as resource,
  case
    when not default_ebs_encryption_enabled then 'alarm'
    else 'ok'
  end as status,
  case
    when not default_ebs_encryption_enabled then region || ' default EBS encryption disabled.'
    else region || ' default EBS encryption enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_regional_settings;