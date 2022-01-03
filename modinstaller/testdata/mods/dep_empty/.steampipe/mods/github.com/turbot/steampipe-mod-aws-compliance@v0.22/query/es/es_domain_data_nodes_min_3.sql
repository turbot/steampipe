select
  -- Required Columns
  arn as resource,
  case
    when elasticsearch_cluster_config ->> 'ZoneAwarenessEnabled' = 'false' then 'alarm'
    when
      elasticsearch_cluster_config ->> 'ZoneAwarenessEnabled' = 'true'
      and (elasticsearch_cluster_config ->> 'InstanceCount')::integer >= 3 then 'ok'
    else 'alarm'
  end status,
  case
    when elasticsearch_cluster_config ->> 'ZoneAwarenessEnabled' = 'false' then title || ' zone awareness disabled.'
    else title || ' has ' || (elasticsearch_cluster_config ->> 'InstanceCount') || ' data node(s).'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;