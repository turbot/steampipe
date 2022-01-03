with bad_rules as (
  select
    network_acl_id,
    count(*) as num_bad_rules
  from
    aws_vpc_network_acl,
    jsonb_array_elements(entries) as att
  where
    att ->> 'Egress' = 'false' -- as per aws egress = false indicates the ingress
    and (
      att ->> 'CidrBlock' = '0.0.0.0/0'
      or att ->> 'Ipv6CidrBlock' =  '::/0'
    )
    and att ->> 'RuleAction' = 'allow'
    and (
      (
        att ->> 'Protocol' = '-1' -- all traffic
        and att ->> 'PortRange' is null
      )
      or (
        (att -> 'PortRange' ->> 'From') :: int <= 22
        and (att -> 'PortRange' ->> 'To') :: int >= 22
        and att ->> 'Protocol' in('6', '17')  -- TCP or UDP
      )
      or (
        (att -> 'PortRange' ->> 'From') :: int <= 3389
        and (att -> 'PortRange' ->> 'To') :: int >= 3389
        and att ->> 'Protocol' in('6', '17')  -- TCP or UDP
    )
  )
  group by
    network_acl_id
)

select
  -- Required Columns
  'arn:' || acl.partition || ':ec2:' || acl.region || ':' || acl.account_id || ':network-acl/' || acl.network_acl_id  as resource,
  case
    when bad_rules.network_acl_id is null then 'ok'
    else 'alarm'
  end as status,
  case
    when bad_rules.network_acl_id is null then acl.network_acl_id || ' does not allow ingress to port 22 or 3389 from 0.0.0.0/0 or ::/0.'
    else acl.network_acl_id || ' contains ' || bad_rules.num_bad_rules || ' rule(s) allowing ingress to port 22 or 3389 from 0.0.0.0/0 or ::/0.'
  end as reason,
  -- Additional Dimensions
  acl.region,
  acl.account_id
from	
  aws_vpc_network_acl as acl
  left join bad_rules on bad_rules.network_acl_id = acl.network_acl_id
