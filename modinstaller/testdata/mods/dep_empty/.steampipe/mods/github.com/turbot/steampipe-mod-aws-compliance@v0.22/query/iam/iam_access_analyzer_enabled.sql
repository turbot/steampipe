select
  -- Required Columns
  r.region as resource,
  case
    when aa.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when aa.arn is not null then aa.name ||  ' enabled in ' || r.region || '.'
    else 'Access analyzer not enabled in ' || r.region || '.'
  end as reason,
  -- Additional Dimensions
  r.region,
  r.account_id
from
  aws_region as r
  left join aws_accessanalyzer_analyzer as aa on r.account_id = aa.account_id and r.region = aa.region
where
  r.opt_in_status != 'not-opted-in';
