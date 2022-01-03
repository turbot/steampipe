with cross_account_buckets as (
  select
    distinct arn
  from
    aws_s3_bucket,
    jsonb_array_elements(policy_std -> 'Statement') as s,
    jsonb_array_elements_text(s -> 'Principal' -> 'AWS') as p,
    string_to_array(p, ':') as pa,
    jsonb_array_elements_text(s -> 'Action') as a
  where
    s ->> 'Effect' = 'Allow'
    and (
      pa [5] != account_id
      or p = '*'
    )
    and a in (
      's3:deletebucketpolicy',
      's3:putbucketacl',
      's3:putbucketpolicy',
      's3:putencryptionconfiguration',
      's3:putobjectacl'
    )
)
select
  -- Required Columns
  a.arn as resource,
  case
    when b.arn is null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.arn is null then title || ' restricts cross-account bucket access.'
    else title || ' allows cross-account bucket access.'
  end as reason,
  -- Additionl Dimensions
  a.region,
  a.account_id
from
  aws_s3_bucket a
  left join cross_account_buckets b on a.arn = b.arn;