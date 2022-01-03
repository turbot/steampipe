with detailed_listeners as (
  select
    arn,
    load_balancer_arn,
    protocol
  from
    aws_ec2_load_balancer_listener,
    jsonb_array_elements(default_actions) as ac
  where
    split_part(arn,'/',2) = 'app'
    and protocol = 'HTTP'
    and ac ->> 'Type' = 'redirect'
    and ac -> 'RedirectConfig' ->> 'Protocol' = 'HTTPS'
)
select
  -- Required Columns
  a.arn as resource,
  case
    when b.load_balancer_arn is null then 'alarm'
    else 'ok'
  end as status,
   case
    when b.load_balancer_arn is not null then  a.title || ' associated with HTTP redirection.'
    else a.title || ' not associated with HTTP redirection.'
  end as reason,
  -- Additional Dimensions
  a.region,
  a.account_id
from
  aws_ec2_application_load_balancer a
  left join detailed_listeners b on a.arn = b.load_balancer_arn;