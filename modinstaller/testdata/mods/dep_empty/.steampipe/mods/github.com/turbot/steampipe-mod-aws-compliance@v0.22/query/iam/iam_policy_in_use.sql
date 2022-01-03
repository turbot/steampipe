with all_attached_policy_arn as (
  select
    user_arn  as arn
  from
    aws_iam_user as p,
    jsonb_array_elements(attached_policy_arns) as user_arn
  union
  select
  group_arn as arn
  from
    aws_iam_group as g,
    jsonb_array_elements(attached_policy_arns) as group_arn
  where
   users is not null
  union
  select
  role_arn  as role_arn
  from
    aws_iam_role as r,
    jsonb_array_elements(attached_policy_arns) as role_arn
),
distinct_attached_policy_arn as (
  select
  distinct arn  as arn
  from
    all_attached_policy_arn
)
select
  -- Required Columns
  p.arn,
  case
    when a.arn is not null
     then 'ok'
    else 'alarm'
  end as status,
  case
    when a.arn is not null
     then title || ' is attached to resource.'
    else  title || ' is not attached to any resource.'
  end  as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_policy as p
  left join distinct_attached_policy_arn
  as a on p.arn = trim('"' FROM a.arn::text);