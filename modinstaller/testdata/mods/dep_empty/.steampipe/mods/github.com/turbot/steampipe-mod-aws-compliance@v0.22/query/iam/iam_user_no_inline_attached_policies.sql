select
  -- Required Columns
  arn as resource,
  case
    when inline_policies is null and attached_policy_arns is null then 'ok'
    else 'alarm'
  end status,
  name || ' has ' || coalesce(jsonb_array_length(inline_policies),0) || ' inline and ' ||
    coalesce(jsonb_array_length(attached_policy_arns),0) || ' directly attached policies.' as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_user;
