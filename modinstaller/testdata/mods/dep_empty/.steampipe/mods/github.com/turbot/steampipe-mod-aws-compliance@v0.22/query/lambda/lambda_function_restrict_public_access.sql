select
  -- Required Columns
  arn as resource,
  case
    when policy_std -> 'Statement' ->> 'Effect' = 'Allow'
    and (
      policy_std -> 'Statement' ->> 'Prinipal' = '*'
      or ( policy_std -> 'Principal' -> 'AWS' ) :: text = '*'
    ) then 'alarm'
    else 'ok'
  end status,
  case
    when policy_std is null then title || ' has no policy.'
    when policy_std -> 'Statement' ->> 'Effect' = 'Allow'
    and (
      policy_std -> 'Statement' ->> 'Prinipal' = '*'
      or ( policy_std -> 'Principal' -> 'AWS' ) :: text = '*'
    ) then title || ' allows public access.'
    else title || ' does not allow public access.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_lambda_function;
