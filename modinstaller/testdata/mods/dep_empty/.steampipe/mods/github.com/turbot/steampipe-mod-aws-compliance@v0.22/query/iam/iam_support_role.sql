-- pgFormatter-ignore
with support_role_count as
(
  select
    -- Required Columns
    'arn:' || a.partition || ':::' || a.account_id as resource,
    count(policy_arn),
    a.account_id
  from
    aws_account as a
    left join aws_iam_role as r on r.account_id = a.account_id
    left join jsonb_array_elements_text(attached_policy_arns) as policy_arn  on true
  where
    split_part(policy_arn, '/', 2) = 'AWSSupportAccess'
    or policy_arn is null
  group by
    a.account_id,
    a.partition
)
select
  -- Required Columns
  resource,
  case
    when count > 0 then 'ok'
    else 'alarm'
  end as status,
  case
    when count = 1 then 'AWSSupportAccess policy attached to 1 role.'
    when count > 1 then 'AWSSupportAccess policy attached to ' || count || ' roles.'
    else 'AWSSupportAccess policy not attached to any role.'
  end  as reason,
  -- Additional Dimensions
  account_id
from
  support_role_count;
