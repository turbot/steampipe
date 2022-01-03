with public_bucket_data as (
-- note the counts are not exactly CORRECT because of the jsonb_array_elements joins,
-- but will be non-zero if any matches are found
  select
    t.s3_bucket_name as name,
    b.arn,
    t.region,
    t.account_id,
    count(acl_grant) filter (where acl_grant -> 'Grantee' ->> 'URI' like '%acs.amazonaws.com/groups/global/AllUsers') as all_user_grants,
    count(acl_grant) filter (where acl_grant -> 'Grantee' ->> 'URI' like '%acs.amazonaws.com/groups/global/AuthenticatedUsers') as auth_user_grants,
    count(s) filter (where s ->> 'Effect' = 'Allow' and  p = '*' ) as anon_statements
  from
    aws_cloudtrail_trail as t
  left join aws_s3_bucket as b on t.s3_bucket_name = b.name
  left join jsonb_array_elements(acl -> 'Grants') as acl_grant on true
  left join jsonb_array_elements(policy_std -> 'Statement') as s  on true
  left join jsonb_array_elements_text(s -> 'Principal' -> 'AWS') as p  on true
  group by
    t.s3_bucket_name,
    b.arn,
    t.region,
    t.account_id
)

select
  -- Required Columns
  case
    when arn is null then 'arn:aws:s3::' || name
    else arn
  end as resource,
  case
    when arn is null then 'skip'
    when all_user_grants > 0 then 'alarm'
    when auth_user_grants > 0 then 'alarm'
    when anon_statements > 0 then 'alarm'
    else 'ok'
  end as status,
  case
    when arn is null then name || ' not found in account ' || account_id || '.'
    when all_user_grants > 0 then name || ' grants access to AllUsers in ACL.'
    when auth_user_grants > 0 then name || ' grants access to AuthenticatedUsers in ACL.'
    when anon_statements > 0 then name || ' grants access to AWS:*" in bucket policy.'
    else name || ' does not grant anonymous access in ACL or bucket policy.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  public_bucket_data
