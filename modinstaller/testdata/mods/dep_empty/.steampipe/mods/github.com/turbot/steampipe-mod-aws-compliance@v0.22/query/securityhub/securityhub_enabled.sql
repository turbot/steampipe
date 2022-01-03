select
  -- Required Columns
  hub_arn as resource,
  case
    when a.region = any (ARRAY ['af-south-1', 'eu-south-1', 'cn-north-1', 'cn-northwest-1', 'ap-northeast-3']) then 'skip'
    when r.hub_arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when a.region = any (ARRAY ['af-south-1', 'eu-south-1', 'cn-north-1', 'cn-northwest-1', 'ap-northeast-3']) then 'Region not supported.'
    when r.hub_arn is not null then r.title || ' enabled.'
    else 'Security Hub disabled in ' || a.region || '.'
  end as reason,
  -- Additional Dimensions
  a.region,
  a.account_id
from
  aws_region as a
  left join aws_securityhub_hub as r on r.region = a.name;