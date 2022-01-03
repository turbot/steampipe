with trails_enabled as (
  select
    distinct arn,
    is_logging,
    event_selectors,
    coalesce(
      jsonb_agg(g) filter ( where not (g = 'null') ),
      $$[]$$::jsonb
    ) as excludeManagementEventSources
  from
    aws_cloudtrail_trail
    left join jsonb_array_elements(event_selectors) as e on true
    left join jsonb_array_elements_text(e -> 'ExcludeManagementEventSources') as g on true
  where
    home_region = region
  group by arn, is_logging, event_selectors
),
all_trails as (
  select
    a.arn as arn,
    case
      when a.is_logging is null then b.is_logging
      else a.is_logging
    end as is_logging,
    case
      when a.event_selectors is null then b.event_selectors
      else a.event_selectors
    end as event_selectors,
    b.excludeManagementEventSources,
    a.include_global_service_events,
    a.is_multi_region_trail,
    a.log_file_validation_enabled,
    a.kms_key_id,
    a.region,
    a.account_id,
    a.title
  from
    aws_cloudtrail_trail as a
    left join trails_enabled as b on a.arn = b.arn
)
select
  -- Required Columns
  arn as resource,
  case
    when not is_logging then 'alarm'
    when not include_global_service_events then 'alarm'
    when not is_multi_region_trail then 'alarm'
    when not log_file_validation_enabled then 'alarm'
    when kms_key_id is null then 'alarm'
    when not (jsonb_array_length(event_selectors) = 1 and event_selectors @> '[{"ReadWriteType":"All"}]') then 'alarm'
    when not (event_selectors @> '[{"IncludeManagementEvents":true}]') then 'alarm'
    when jsonb_array_length(excludeManagementEventSources) > 0 then 'alarm'
    else 'ok'
  end as status,
  case
    when not is_logging then title || ' disabled.'
    when not include_global_service_events then title || ' not recording global service events.'
    when not is_multi_region_trail then title || ' not a muti-region trail.'
    when not log_file_validation_enabled then title || ' log file validation disabled.'
    when kms_key_id is null then title || ' not encrypted with a KMS key.'
    when not (jsonb_array_length(event_selectors) = 1 and event_selectors @> '[{"ReadWriteType":"All"}]') then title || ' not recording events for both reads and writes.'
    when not (event_selectors @> '[{"IncludeManagementEvents":true}]') then title || ' not recording management events.'
    when jsonb_array_length(excludeManagementEventSources) > 0 then title || ' excludes management events for ' || trim(excludeManagementEventSources::text, '[]') || '.'
    else title || ' meets all security best practices.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  all_trails;
