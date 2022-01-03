select
  -- Required Columns
  launch_configuration_arn as resource,
  case
    when associate_public_ip_address then 'alarm'
    else 'ok'
  end as status,
  case
    when associate_public_ip_address then title || ' public IP enabled.'
    else title || ' public IP disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_launch_configuration;