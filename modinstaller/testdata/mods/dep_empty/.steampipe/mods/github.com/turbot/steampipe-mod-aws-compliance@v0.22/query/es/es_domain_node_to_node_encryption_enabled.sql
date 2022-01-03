select
  -- Required Columns
  arn as resource,
  case
    when not enabled then 'alarm'
    else 'ok'
  end as status,
  case
    when not enabled then title || ' node-to-node encryption disabled.'
    else title || ' node-to-node encryption enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain
where
  region != any (ARRAY ['af-south-1', 'eu-south-1', 'cn-north-1', 'cn-northwest-1']);