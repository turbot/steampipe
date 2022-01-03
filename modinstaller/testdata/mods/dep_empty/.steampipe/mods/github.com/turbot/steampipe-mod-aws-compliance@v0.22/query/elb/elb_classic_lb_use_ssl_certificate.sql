with detailed_classic_listeners as (
  select
    name
  from
    aws_ec2_classic_load_balancer,
    jsonb_array_elements(listener_descriptions) as listener_description
  where
    listener_description -> 'Listener' ->> 'Protocol' in ('HTTPS', 'SSL', 'TLS')
    and listener_description -> 'Listener' ->> 'SSLCertificateId' like 'arn:aws:acm%'
)
select
  -- Required Columns
  'arn:' || a.partition || ':elasticloadbalancing:' || a.region || ':' || a.account_id || ':loadbalancer/' || a.name as resource,
  case
    when a.listener_descriptions is null then 'skip'
    when b.name is not null then 'alarm'
    else 'ok'
  end as status,
  case
    when a.listener_descriptions is null then a.title || ' has no listener.'
    when b.name is not null then a.title || ' does not use certificates provided by ACM.'
    else a.title || ' uses certificates provided by ACM.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ec2_classic_load_balancer as a
  left join detailed_classic_listeners as b on a.name = b.name;