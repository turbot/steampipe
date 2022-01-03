with ingress_ssh_rules as (
  select
    group_id,
    count(*) as num_ssh_rules
  from
    aws_vpc_security_group_rule
  where
    type = 'ingress'
    and cidr_ip = '0.0.0.0/0'
    and (
        ( ip_protocol = '-1'
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
        or (
            from_port >= 21
            and to_port <= 21
        )
        or (
            from_port >= 20
            and to_port <= 20
        )
        or (
            from_port >= 3306
            and to_port <= 3306
        )
        or (
            from_port >= 4333
            and to_port <= 4333
        )
    )
  group by
    group_id
)
select
  -- Required Columns
  arn as resource,
  case
    when ingress_ssh_rules.group_id is null then 'ok'
    else 'alarm'
  end as status,
  case
    when ingress_ssh_rules.group_id is null then sg.group_id || ' ingress restricted for ports 20, 21, 22, 3306, 3389, 4333 from 0.0.0.0/0.'
    else  sg.group_id || ' contains ' || ingress_ssh_rules.num_ssh_rules || ' ingress rule(s) allowing access on ports 20, 21, 22, 3306, 3389, 4333 from 0.0.0.0/0.'
  end as reason,
  -- Additional Dimensions
  sg.region,
  sg.account_id
from
  aws_vpc_security_group as sg
  left join ingress_ssh_rules on ingress_ssh_rules.group_id = sg.group_id;
