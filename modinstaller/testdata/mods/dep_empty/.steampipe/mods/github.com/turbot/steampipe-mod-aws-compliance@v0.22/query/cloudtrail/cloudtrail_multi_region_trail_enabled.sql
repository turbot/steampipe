with multi_region_trails as (
  select
    account_id,
    count(account_id) as num_multregion_trails
  from
    aws_cloudtrail_trail
  where
    is_multi_region_trail
    and is_logging
  group by
    account_id,
    is_multi_region_trail
)
select
  -- Required Columns
  a.arn as resource,
  case
    when coalesce(num_multregion_trails, 0)  < 1 then 'alarm'
    else 'ok'
  end as status,
  a.title || ' has ' || coalesce(num_multregion_trails, 0) || ' multi-region trail(s).' as reason,
  -- Additional Dimensions
  a.account_id
from
  aws_account as a
left join multi_region_trails as b on a.account_id = b.account_id;