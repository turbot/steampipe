select
  -- Required Columns
  arn as resource,
  case
    when public_ip_address is null then 'ok'
    else 'alarm'
  end status,
  case
    when public_ip_address is null then instance_id || ' not publicly accessible.'
    else instance_id || ' publicly accessible.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;