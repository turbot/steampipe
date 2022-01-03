select
  -- Required Columns
  arn as resource,
  case
    when load_balancer_attributes @> '[{"Key": "deletion_protection.enabled", "Value": "true"}]' then 'ok'
    else 'alarm'
  end as status,
  case
    when load_balancer_attributes @> '[{"Key": "deletion_protection.enabled", "Value": "true"}]' then title || ' deletion protection enabled.'
    else title || ' deletion protection disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_application_load_balancer;