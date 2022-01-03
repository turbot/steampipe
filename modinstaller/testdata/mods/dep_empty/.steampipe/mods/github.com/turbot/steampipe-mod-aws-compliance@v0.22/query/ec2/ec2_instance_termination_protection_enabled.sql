select
  -- Required Columns
  arn as resource,
  case
    when disable_api_termination then 'ok'
    else 'alarm'
  end status,
  case
    when disable_api_termination then instance_id || ' termination protection enabled.'
    else instance_id || ' termination protection disabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;