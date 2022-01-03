with trails_enabled as (
  select
    arn,
    is_logging
  from
    aws_cloudtrail_trail
  where
    home_region = region
)
select
  -- Required Columns
  a.arn as resource,
  case
    when b.is_logging is null and a.is_logging then 'ok'
    when b.is_logging then 'ok'
    else 'alarm'
  end as status,
  case
    when b.is_logging is null and a.is_logging then a.title || ' enabled.'
    when b.is_logging then a.title || ' enabled.'
    else a.title || ' disabled.'
  end as reason,
  -- Additional Dimensions
  a.region,
  a.account_id
from
  aws_cloudtrail_trail as a
left join trails_enabled b on a.arn = b.arn;