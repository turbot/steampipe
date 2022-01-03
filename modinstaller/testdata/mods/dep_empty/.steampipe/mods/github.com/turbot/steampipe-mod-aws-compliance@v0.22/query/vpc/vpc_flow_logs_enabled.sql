with vpc_with_active_flow_logs as (
  select
    v.akas ->> 0 as resource,
    count(f.resource_id),
    v.vpc_id,
    f.flow_log_status,
    v.region,
    v.account_id,
    f.resource_id
  from
    aws_vpc as v
    left join aws_vpc_flow_log as f on v.vpc_id = f.resource_id
  group by
    resource,
    v.vpc_id,
    f.flow_log_status,
    v.region,
    v.account_id,
    f.resource_id
)
select
  -- Required columns
  resource,
  case
    when count > 0 then 'ok'
    else 'alarm'
  end as status,
  case
    when count > 0 then vpc_id || ' flow logging enabled.'
    else vpc_id || ' flow logging disabled.'
  end as reason,
  -- Additional columns
  region,
  account_id
from
  vpc_with_active_flow_logs
group by
  vpc_id,
  count,
  flow_log_status,
  resource,
  region,
  account_id;
