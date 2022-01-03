select
  -- Required Columns
  arn as resource,
  case
    when monitoring_state = 'enabled' then 'ok'
    else 'alarm'
  end status,
  case
    when monitoring_state = 'enabled' then instance_id || ' detailed monitoring enabled.'
    else instance_id || ' detailed monitoring disabled.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;