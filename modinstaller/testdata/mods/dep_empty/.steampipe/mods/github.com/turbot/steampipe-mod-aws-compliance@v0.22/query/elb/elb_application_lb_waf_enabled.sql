select
  -- Required Columns
  arn as resource,
  case
    when load_balancer_attributes @> '[{"Key":"waf.fail_open.enabled","Value":"true"}]' then 'ok'
    else 'alarm'
  end as status,
  case
    when load_balancer_attributes @> '[{"Key":"waf.fail_open.enabled","Value":"true"}]' then title || ' WAF enabled.'
    else title || ' WAF disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_application_load_balancer;