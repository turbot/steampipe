
select
  -- Required Columns
  u.arn as resource,
  case
    when count(k.*) > 1 then 'alarm'
    else 'ok'
  end as status,
  u.name || ' has ' || count(k.*) || ' active access keys.' as reason,
  -- Additional Dimensions
  u.account_id
from aws_iam_user as u
left join aws_iam_access_key as k on u.name = k.user_name and u.account_id = u.account_id
where
  k.status = 'Active' or k.status is null
group by
  u.arn,
  u.name,
  u.account_id
