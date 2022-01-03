 with trail_details as (
  select
    name as trail_name,
    arn,
    is_multi_region_trail,
    is_logging,
    event_selectors,
    e ->> 'ReadWriteType' as read_write_type,
    account_id,
    region
  from
    aws_cloudtrail_trail,
    jsonb_array_elements(event_selectors) as e
)
select
  -- Required Columns
  arn as resource,
  case
    when not trail_details.is_multi_region_trail then 'alarm'
    when not trail_details.is_logging then 'alarm'
    when read_write_type <> 'All' then 'alarm'
    else 'ok'
  end as status,
  trail_details.trail_name ||
    case when trail_details.is_multi_region_trail then ' is ' else ' is not ' end || 'multi-region,' ||
    case when trail_details.is_logging then ' logging enabled' else ' logging disabled' end ||
    ' for ' || read_write_type || ' events.'
  as reason,
  -- Additional Dimensions
  region,
  account_id
from
  trail_details
