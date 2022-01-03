with s3_selectors as
(
  select
    name as trail_name,
    is_multi_region_trail,
    bucket_selector
  from
    aws_cloudtrail_trail,
    jsonb_array_elements(event_selectors) as event_selector,
    jsonb_array_elements(event_selector -> 'DataResources') as data_resource,
    jsonb_array_elements_text(data_resource -> 'Values') as bucket_selector
  where
    is_multi_region_trail
    and data_resource ->> 'Type' = 'AWS::S3::Object'
    and event_selector ->> 'ReadWriteType' = 'All'
)
select
  -- Required columns
  b.arn as resource,
  case
    when count(bucket_selector) > 0 then 'ok'
    else 'alarm'
  end as status,
  case
    when count(bucket_selector) > 0 then b.name || ' object-level data events logging enabled.'
    else b.name || ' object-level data events logging disabled.'
  end as reason,
  -- Additional columns
  region,
  account_id
from
  aws_s3_bucket as b
  left join
    s3_selectors
    on bucket_selector like (b.arn || '%')
    or bucket_selector = 'arn:aws:s3'
group by
  b.account_id, b.region, b.arn, b.name;