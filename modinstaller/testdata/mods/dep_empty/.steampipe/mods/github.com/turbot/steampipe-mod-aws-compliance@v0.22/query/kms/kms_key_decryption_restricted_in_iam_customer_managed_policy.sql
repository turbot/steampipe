with policy_with_decrypt_grant as (
  select
    distinct arn
  from
    aws_iam_policy,
    jsonb_array_elements(policy_std -> 'Statement') as statement
  where
    not is_aws_managed
    and statement ->> 'Effect' = 'Allow'
    and statement -> 'Resource' ?| array['*', 'arn:aws:kms:*:' || account_id || ':key/*', 'arn:aws:kms:*:' || account_id || ':alias/*']
    and statement -> 'Action' ?| array['*', 'kms:*', 'kms:decrypt', 'kms:reencryptfrom', 'kms:reencrypt*']
)
select
  -- Required Columns
  i.arn as resource,
  case
    when d.arn is null then 'ok'
    else 'alarm'
  end as status,
  case
    when d.arn is null then i.title || ' doesn''t allow decryption actions on all keys.'
    else i.title || ' allows decryption actions on all keys.'
  end as reason,
  -- Additional Dimensions
  i.account_id
from
  aws_iam_policy i
left join policy_with_decrypt_grant d on i.arn = d.arn
where
  not is_aws_managed;