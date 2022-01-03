select
  -- Required Columns
  arn as resource,
  case
    when jsonb_array_length(network_interfaces) = 1 then 'ok'
    else 'alarm'
  end status,
  title || ' has ' || jsonb_array_length(network_interfaces) || ' ENI(s) attached.'
  as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_instance;