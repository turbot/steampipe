with ssl_ok as (
  select
    distinct name,
    arn,
    'ok' as status
  from
    aws_s3_bucket,
    jsonb_array_elements(policy_std -> 'Statement') as s,
    jsonb_array_elements_text(s -> 'Principal' -> 'AWS') as p,
    jsonb_array_elements_text(s -> 'Action') as a,
    jsonb_array_elements_text(s -> 'Resource') as r,
    jsonb_array_elements_text(
      s -> 'Condition' -> 'Bool' -> 'aws:securetransport'
    ) as ssl
  where
    p = '*'
    and s ->> 'Effect' = 'Deny'
    and ssl :: bool = false
)
select
  -- Required Columns
  b.arn as resource,
  case
    when ok.status = 'ok' then 'ok'
    else 'alarm'
  end status,
  case
    when ok.status = 'ok' then b.name || ' bucket policy enforces HTTPS.'
    else b.name || ' bucket policy does not enforce HTTPS.'
  end reason,
  -- Additional Dimensions
  b.region,
  b.account_id
from
  aws_s3_bucket as b
  left join ssl_ok as ok on ok.name = b.name;