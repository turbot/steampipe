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
        or (
            from_port >= 23
            and to_port <= 23
        ) 
        or (
            from_port >= 25
            and to_port <= 25
        ) 
        or (
            from_port >= 445
            and to_port <= 445
        ) 
        or (
            from_port >= 110
            and to_port <= 110
        )
        or (
            from_port >= 135
            and to_port <= 135
        )
        or (
            from_port >= 143
            and to_port <= 143
        )
        or (
            from_port >= 1433
            and to_port <= 3389
        )
        or (
            from_port >= 3389
            and to_port <= 1434
        )
        or (
            from_port >= 5432
            and to_port <= 5432
        )
        or (
            from_port >= 5500
            and to_port <= 5500
        )
        or (
            from_port >= 5601
            and to_port <= 5601
        )    
        or (
            from_port >= 9200
            and to_port <= 9300
        )    
        or (
            from_port >= 8080
            and to_port <= 8080
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
    when ingress_ssh_rules.group_id is null then sg.group_id || ' ingress restricted for common ports from 0.0.0.0/0..'
    else  sg.group_id || ' contains ' || ingress_ssh_rules.num_ssh_rules || ' ingress rule(s) allowing access for common ports from 0.0.0.0/0.'
  end as reason,
  -- Additional Dimensions
  sg.region,
  sg.account_id
from
  aws_vpc_security_group as sg
  left join ingress_ssh_rules on ingress_ssh_rules.group_id = sg.group_id;
