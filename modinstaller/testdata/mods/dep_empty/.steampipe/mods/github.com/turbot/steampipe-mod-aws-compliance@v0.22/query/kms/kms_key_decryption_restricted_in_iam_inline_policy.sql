with user_with_decrypt_grant as (
  select
    distinct arn
  from
    aws_iam_user,
    jsonb_array_elements(inline_policies_std) as inline_policy,
    jsonb_array_elements(inline_policy -> 'PolicyDocument' -> 'Statement') as statement
  where
    statement ->> 'Effect' = 'Allow'
    and statement -> 'Resource' ?| array['*', 'arn:aws:kms:*:' || account_id || ':key/*', 'arn:aws:kms:*:' || account_id || ':alias/*']
    and statement -> 'Action' ?| array['*', 'kms:*', 'kms:decrypt', 'kms:deencrypt*', 'kms:reencryptfrom']
),
role_with_decrypt_grant as (
  select
    distinct arn
  from
    aws_iam_role,
    jsonb_array_elements(inline_policies_std) as inline_policy,
    jsonb_array_elements(inline_policy -> 'PolicyDocument' -> 'Statement') as statement
  where
    statement ->> 'Effect' = 'Allow'
    and statement -> 'Resource' ?| array['*', 'arn:aws:kms:*:' || account_id || ':key/*', 'arn:aws:kms:*:' || account_id || ':alias/*']
    and statement -> 'Action' ?| array['*', 'kms:*', 'kms:decrypt', 'kms:deencrypt*', 'kms:reencryptfrom']
),
group_with_decrypt_grant as (
  select
    distinct arn
  from
    aws_iam_group,
    jsonb_array_elements(inline_policies_std) as inline_policy,
    jsonb_array_elements(inline_policy -> 'PolicyDocument' -> 'Statement') as statement
  where
    statement ->> 'Effect' = 'Allow'
    and statement -> 'Resource' ?| array['*', 'arn:aws:kms:*:' || account_id || ':key/*', 'arn:aws:kms:*:' || account_id || ':alias/*']
    and statement -> 'Action' ?| array['*', 'kms:*', 'kms:decrypt', 'kms:deencrypt*', 'kms:reencryptfrom']
)
select
  -- Required Columns
  i.arn as resource,
  case
    when d.arn is null then 'ok'
    else 'alarm'
  end as status,
  case
    when d.arn is null then 'User ' || i.title || ' not allowed to perform decryption actions on all keys.'
    else 'User ' || i.title || ' allowed to perform decryption actions on all keys.'
  end as reason,
  -- Additional Dimensions
  i.account_id
from
  aws_iam_user i
  left join user_with_decrypt_grant d on i.arn = d.arn
union
select
  -- Required Columns
  r.arn as resource,
  case
    when d.arn is null then 'ok'
    else 'alarm'
  end as status,
  case
    when d.arn is null then 'Role ' || r.title || ' not allowed to perform decryption actions on all keys.'
    else 'Role ' || r.title || ' allowed to perform decryption actions on all keys.'
  end as reason,
  -- Additional Dimensions
  r.account_id
from
  aws_iam_role r
  left join role_with_decrypt_grant d on r.arn = d.arn
where
  r.arn not like '%service-role/%'
union
select
  -- Required Columns
  g.arn as resource,
  case
    when d.arn is null then 'ok'
    else 'alarm'
  end as status,
  case
    when d.arn is null then 'Role ' || g.title || ' not allowed to perform decryption actions on all keys.'
    else 'Group ' || g.title || ' allowed to perform decryption actions on all keys.'
  end as reason,
  -- Additional Dimensions
  g.account_id
from
  aws_iam_group g
  left join group_with_decrypt_grant d on g.arn = d.arn;