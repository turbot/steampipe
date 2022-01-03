-- pgFormatter-ignore
-- Get count for any region with all matching criteria
with global_recorders as (
  select
    count(*) as global_config_recorders
  from
    aws_config_configuration_recorder
  where
    recording_group -> 'IncludeGlobalResourceTypes' = 'true'
    and recording_group -> 'AllSupported' = 'true'
    and status ->> 'Recording' = 'true'
    and status ->> 'LastStatus' = 'SUCCESS'
 )
select
  -- Required columns
  'arn:aws::' || a.region || ':' || a.account_id as resource,
  case
  -- When any of the region satisfies with above CTE
  -- In left join of <aws_config_configuration_recorder> table, regions now having
  -- 'Recording' and 'LastStatus' matching criteria can be considered as OK
    when
      g.global_config_recorders >= 1
      and status ->> 'Recording' = 'true'
      and status ->> 'LastStatus' = 'SUCCESS'
    then 'ok'
    else 'alarm'
  end as status,
  -- Below cases are for citing respective reasons for control state
  case
    when recording_group -> 'IncludeGlobalResourceTypes' = 'true' then a.region || ' IncludeGlobalResourceTypes enabled,'
    else a.region || ' IncludeGlobalResourceTypes disabled,'
  end ||
  case
    when recording_group -> 'AllSupported' = 'true' then ' AllSupported enabled,'
    else ' AllSupported disabled,'
  end ||
  case
    when status ->> 'Recording' = 'true' then ' Recording enabled'
    else ' Recording disabled'
  end ||
  case
    when status ->> 'LastStatus' = 'SUCCESS' then ' and LastStatus is SUCCESS.'
    else ' and LastStatus is not SUCCESS.'
  end as reason,
  -- Additional columns
  a.region,
  a.account_id
from
  global_recorders as g,
  aws_region as a
  left join aws_config_configuration_recorder as r
    on r.account_id = a.account_id and r.region = a.name;
