with host_network_task_definition as (
  select
    distinct task_definition_arn as arn
  from
    aws_ecs_task_definition,
    jsonb_array_elements(container_definitions) as c
  where
    network_mode = 'host'
    and
      (c ->> 'Privileged' is not null
        and c ->> 'Privileged' <> 'false'
      )
    and
      ( c ->> 'User' is not null
      and c ->> 'User' <> 'root'
      )
)
select
  -- Required Columns
  a.task_definition_arn as resource,
  case
    when a.network_mode is null or a.network_mode <> 'host' then 'skip'
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when a.network_mode is null or a.network_mode <> 'host' then a.title || ' not host network mode.'
    when b.arn is not null then a.title || ' have secure host network mode.'
    else a.title || ' not have secure host network mode.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ecs_task_definition as a
  left join host_network_task_definition as b on a.task_definition_arn = b.arn;