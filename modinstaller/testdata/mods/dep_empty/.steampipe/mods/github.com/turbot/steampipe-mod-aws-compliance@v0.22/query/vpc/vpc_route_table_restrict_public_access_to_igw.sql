with route_with_public_access as (
  select
    route_table_id,
    count(*) as num
  from
    aws_vpc_route_table,
    jsonb_array_elements(routes) as r
  where
    ( r ->> 'DestinationCidrBlock' = '0.0.0.0/0'
      or r ->> 'DestinationCidrBlock' = '::/0'
    )
    and r ->> 'GatewayId' like 'igw%'
  group by
    route_table_id
)
select
  -- Required Columns
  a.route_table_id as resource,
  case
    when b.route_table_id is null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.route_table_id is null then a.title || ' does not have public routes to an Internet Gateway (IGW)'
    else a.title || ' contains ' || b.num || ' rule(s) which have public routes to an Internet Gateway (IGW)'
  end as reason,
  -- Additional Dimensions
  a.region,
  a.account_id
from
  aws_vpc_route_table as a
  left join route_with_public_access as b on b.route_table_id = a.route_table_id
