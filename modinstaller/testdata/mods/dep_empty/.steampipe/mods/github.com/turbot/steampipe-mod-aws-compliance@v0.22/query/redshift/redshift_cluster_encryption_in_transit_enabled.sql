with pg_with_ssl as (
 select
  name as pg_name,
  p ->> 'ParameterName' as parameter_name,
  p ->> 'ParameterValue' as parameter_value
from
  aws_redshift_parameter_group,
  jsonb_array_elements(parameters) as p
where
  p ->> 'ParameterName' = 'require_ssl'
  and p ->> 'ParameterValue' = 'true'
)
select
  -- Required Columns
  'arn:aws:redshift:' || region || ':' || account_id || ':' || 'cluster' || ':' || cluster_identifier as resource,
  case
    when cpg ->> 'ParameterGroupName' in (select pg_name from pg_with_ssl ) then 'ok'
    else 'alarm'
  end as status,
  case
    when cpg ->> 'ParameterGroupName' in (select pg_name from pg_with_ssl ) then title || ' encryption in transit enabled.'
    else title || ' encryption in transit disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_redshift_cluster,
  jsonb_array_elements(cluster_parameter_groups) as cpg;