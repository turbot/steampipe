select
  -- Required Columns
  'arn:' || partition || ':elasticloadbalancing:' || region || ':' || account_id || ':loadbalancer/' || title as resource,
  case
    when listener_description -> 'Listener' ->> 'Protocol' in ('HTTPS', 'SSL', 'TLS') then 'ok'
    else 'alarm'
  end as status,
  case
    when listener_description -> 'Listener' ->> 'Protocol' = 'HTTPS' then title || ' configured with HTTPS protocol.'
    when listener_description -> 'Listener' ->> 'Protocol' = 'SSL' then title || ' configured with TLS protocol.'
    else title || ' configured with ' || (listener_description -> 'Listener' ->> 'Protocol') || ' protocol.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_classic_load_balancer,
  jsonb_array_elements(listener_descriptions) as listener_description;