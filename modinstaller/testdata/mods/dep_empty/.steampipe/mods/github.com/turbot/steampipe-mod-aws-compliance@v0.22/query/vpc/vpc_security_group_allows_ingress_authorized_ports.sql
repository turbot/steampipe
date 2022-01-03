with ingress_unauthorized_ports as (
  select
    group_id,
    count(*)
  from
    aws_vpc_security_group_rule
  where
    type = 'ingress'
    and cidr_ip = '0.0.0.0/0'
    and (from_port is null or from_port not in (80,443))
  group by group_id
)
select
  -- Required Columns
  sg.arn as resource,
  case
    when ingress_unauthorized_ports.count > 0 then 'alarm'
    else 'ok'
  end as status,
  case
    when ingress_unauthorized_ports.count > 0 then sg.title || ' having unrestricted incoming traffic other than default ports from 0.0.0.0/0 '
    else sg.title || ' allows unrestricted incoming traffic for authorized default ports (80,443).'
  end as reason,
  sg.region,
  sg.account_id
from
  aws_vpc_security_group as sg
  left join ingress_unauthorized_ports on ingress_unauthorized_ports.group_id = sg.group_id;