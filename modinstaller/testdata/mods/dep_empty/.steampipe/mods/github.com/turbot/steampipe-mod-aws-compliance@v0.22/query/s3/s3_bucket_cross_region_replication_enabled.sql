with bucket_with_replication as (
  select
    name,
    r ->> 'Status' as rep_status
  from
    aws_s3_bucket,
    jsonb_array_elements(replication -> 'Rules' ) as r
)
select
  -- Required Columns
  b.arn as resource,
  case
    when b.name = r.name and r.rep_status = 'Enabled' then 'ok'
    else 'alarm'
  end as status,
  case
    when b.name = r.name and r.rep_status = 'Enabled' then b.title || ' enabled with cross-region replication.'
    else b.title || ' not enabled with cross-region replication.'
  end as reason,
  -- Additional Dimensions
  b.region,
  b.account_id
from
  aws_s3_bucket b
  left join bucket_with_replication r on b.name = r.name;