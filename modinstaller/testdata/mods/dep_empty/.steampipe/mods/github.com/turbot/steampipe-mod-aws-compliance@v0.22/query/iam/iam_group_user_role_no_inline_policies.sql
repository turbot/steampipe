select
  -- Required Columns
  arn as resource,
  case
    when inline_policies is null then 'ok'
    else 'alarm'
  end status,
  'User ' || title || ' has ' || coalesce(jsonb_array_length(inline_policies), 0) || ' inline policies.' as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_user
union
select
  -- Required Columns
  arn as resource,
  case
    when inline_policies is null then 'ok'
    else 'alarm'
  end status,
  'Role ' || title || ' has ' || coalesce(jsonb_array_length(inline_policies), 0) || ' inline policies.' as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_role
where
  arn not like '%service-role/%'
union
select
  -- Required Columns
  arn as resource,
  case
    when inline_policies is null then 'ok'
    else 'alarm'
  end status,
  'Group ' || title || ' has ' || coalesce(jsonb_array_length(inline_policies), 0) || ' inline policies.' as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_group;