select
  -- Required columns
  t.arn as resource,
  case
    when b.logging is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.logging is not null then t.title || '''s logging bucket ' || t.s3_bucket_name || ' has access logging enabled.'
    else t.title || '''s logging bucket ' || t.s3_bucket_name || ' has access logging disabled.'
  end as reason,
  -- Additional columns
  t.region,
  t.account_id
from
  aws_cloudtrail_trail t
  inner join aws_s3_bucket b on t.s3_bucket_name = b.name
where 
  t.region = t.home_region;