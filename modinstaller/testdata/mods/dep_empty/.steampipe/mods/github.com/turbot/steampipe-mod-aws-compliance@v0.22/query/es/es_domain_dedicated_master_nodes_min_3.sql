select
  -- Required Columns
  arn as resource,
  case
    when elasticsearch_cluster_config ->> 'DedicatedMasterEnabled' = 'false' then 'alarm'
    when
      elasticsearch_cluster_config ->> 'DedicatedMasterEnabled' = 'true'
      and (elasticsearch_cluster_config ->> 'DedicatedMasterCount')::integer >= 3 then 'ok'
    else 'alarm'
  end status,
  case
    when elasticsearch_cluster_config ->> 'DedicatedMasterEnabled' = 'false' then title || ' dedicated master nodes disabled.'
    else title || ' has ' || (elasticsearch_cluster_config ->> 'DedicatedMasterCount') || ' dedicated master node(s).'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;