with wildcard_action_policies as (
  select
    arn,
    count(*) as statements_num
  from
    aws_iam_policy,
    jsonb_array_elements(policy_std -> 'Statement') as s,
    jsonb_array_elements_text(s -> 'Resource') as resource,
    jsonb_array_elements_text(s -> 'Action') as action
  where
    is_aws_managed = 'false'
    and s ->> 'Effect' = 'Allow'
    and resource = '*'
    and action like '%:*'
  group by
    arn
)
select
  -- Required Columns
  a.arn as resource,
  case
    when b.arn is null then 'ok'
    else 'alarm'
  end status,
  a.name || ' contains ' || coalesce(b.statements_num,0)  ||
     ' statements that allow action "Service:*" on resource "*".' as reason,
  -- Additional Dimensions
  a.account_id
from
  aws_iam_policy as a
  left join wildcard_action_policies as b on a.arn = b.arn;