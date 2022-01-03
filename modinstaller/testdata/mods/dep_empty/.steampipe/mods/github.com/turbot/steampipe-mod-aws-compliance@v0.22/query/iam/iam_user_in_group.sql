select
  -- Required Columns
  arn as resource,
  case
    when groups is null then 'alarm'
    else 'ok'
  end as status,
  case
    when groups is null then title || ' not associated with any IAM group.'
    else title || ' associated with IAM group.'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_iam_user;