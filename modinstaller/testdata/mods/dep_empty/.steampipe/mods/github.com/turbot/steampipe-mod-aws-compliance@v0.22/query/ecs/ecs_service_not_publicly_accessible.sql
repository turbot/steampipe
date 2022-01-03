with service_awsvpc_mode_task_definition as (
  select
    a.service_name as service_name,
    b.task_definition_arn as task_definition
  from
    aws_ecs_service as a
    left join aws_ecs_task_definition as b on a.task_definition = b.task_definition_arn
  where
    b.network_mode = 'awsvpc'
)
select
  -- Required Columns
  a.arn as resource,
  case
    when b.service_name is null then 'skip'
    when network_configuration -> 'AwsvpcConfiguration' ->> 'AssignPublicIp' = 'DISABLED' then 'ok'
    else 'alarm'
  end as status,
  case
    when b.service_name is null then a.title || ' task definition not host network mode.'
    when network_configuration -> 'AwsvpcConfiguration' ->> 'AssignPublicIp' = 'DISABLED' then a.title || ' not publicly accessible.'
    else a.title || ' publicly accessible.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ecs_service as a
  left join service_awsvpc_mode_task_definition as b on a.service_name = b.service_name;