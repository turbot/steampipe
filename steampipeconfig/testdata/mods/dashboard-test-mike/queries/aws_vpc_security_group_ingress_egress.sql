with analysis as (
  select
    null as parent,
    concat('source-', text(cidr_ip), pair_group_id) as id,
    concat(
      coalesce(text(cidr_ip), pair_group_id),
      ' (source)'
    ) as name,
    0 as depth,
    'aws_vpc_security_group_rule' as category
  from
    aws_vpc_security_group_rule
  where
    group_id = 'sg-0d92bde3dd654d88e'
    and type = 'ingress'
  union
  select
    concat('source-', text(cidr_ip), pair_group_id) as parent,
    concat('ingress-', text(cidr_ip), pair_group_id) as id,
    case
      when ip_protocol = '-1' then 'all (ingress)'
      when ip_protocol = 'icmp' then 'icmp (ingress)'
      when from_port is not null
      and to_port is not null
      and from_port = to_port then concat(from_port, '/', ip_protocol, ' (ingress)')
      else concat(
        from_port,
        '-',
        to_port,
        '/',
        ip_protocol,
        ' (ingress)'
      )
    end as name,
    1 as depth,
    'aws_vpc_security_group_rule' as category
  from
    aws_vpc_security_group_rule
  where
    group_id = 'sg-0d92bde3dd654d88e'
    and type = 'ingress'
  union
  select
    concat('ingress-', text(cidr_ip), pair_group_id) as parent,
    sg.group_id as id,
    sg.group_name as name,
    2 as depth,
    'aws_vpc_security_group' as category
  from
    aws_vpc_security_group sg
    inner join aws_vpc_security_group_rule sgr on sg.group_id = sgr.group_id
  where
    sg.group_id = 'sg-0d92bde3dd654d88e'
  union
  select
    group_id as parent,
    concat('egress-', text(cidr_ip), pair_group_id) as id,
    case
      when ip_protocol = '-1' then 'all (egress)'
      when ip_protocol = 'icmp' then 'icmp (egress)'
      when from_port is not null
      and to_port is not null
      and from_port = to_port then concat(from_port, '/', ip_protocol, ' (egress)')
      else concat(
        from_port,
        '-',
        to_port,
        '/',
        ip_protocol,
        ' (egress)'
      )
    end as name,
    3 as depth,
    'aws_vpc_security_group_rule' as category
  from
    aws_vpc_security_group_rule
  where
    group_id = 'sg-0d92bde3dd654d88e'
    and type = 'egress'
  union
  select
    concat('egress-', text(cidr_ip), pair_group_id) as parent,
    concat('desintation-', text(cidr_ip), pair_group_id) as id,
    concat(
      coalesce(text(cidr_ip), pair_group_id),
      ' (destination)'
    ) as name,
    4 as depth,
    'aws_vpc_security_group_rule' as category
  from
    aws_vpc_security_group_rule
  where
    group_id = 'sg-0d92bde3dd654d88e'
    and type = 'egress'
)
select
  parent,
  id,
  name
  category
from
  analysis
order by
  depth,
  category,
  id;