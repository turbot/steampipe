with bad_rules as (
  select
    group_id,
    count(*) as num_bad_rules
  from
    aws_vpc_security_group_rule
  where
    type = 'ingress'
    and (
      cidr_ip = '0.0.0.0/0'
      or cidr_ipv6 = '::/0'
    )
    and (
        ( ip_protocol = '-1'      -- all traffic
        and from_port is null
        )
        or (
            from_port >= 22
            and to_port <= 22
        )
        or (
            from_port >= 3389
            and to_port <= 3389
        )
    )
  group by
    group_id
)
select
  -- Required Columns
  arn as resource,
  case
    when bad_rules.group_id is null then 'ok'
    else 'alarm'
  end as status,
  case
    when bad_rules.group_id is null then sg.group_id || ' does not allow ingress to port 22 or 3389 from 0.0.0.0/0 or ::/0.'
    else  sg.group_id || ' contains ' || bad_rules.num_bad_rules || ' rule(s) that allow ingress to port 22 or 3389 from 0.0.0.0/0 or ::/0.'
  end as reason,
  -- Additional Dimensions
  sg.region,
  sg.account_id
from	
  aws_vpc_security_group as sg
  left join bad_rules on bad_rules.group_id = sg.group_id
