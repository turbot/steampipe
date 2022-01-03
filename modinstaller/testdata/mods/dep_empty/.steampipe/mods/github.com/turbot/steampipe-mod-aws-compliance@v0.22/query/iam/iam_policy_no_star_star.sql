with bad_policies as (
  select
    arn,
    count(*) as num_bad_statements
  from
    aws_iam_policy,
    jsonb_array_elements(policy_std -> 'Statement') as s,
    jsonb_array_elements_text(s -> 'Resource') as resource,
    jsonb_array_elements_text(s -> 'Action') as action
  where
    s ->> 'Effect' = 'Allow'
    and resource = '*'
    and (
      (action = '*'
      or action = '*:*'
      )
  )
  group by
    arn
)
select
  -- Required Columns
  p.arn as resource,
  case
    when bad.arn is null then 'ok'
    else 'alarm'
  end status,
  p.name || ' contains ' || coalesce(bad.num_bad_statements,0)  ||
     ' statements that allow action "*" on resource "*".' as reason,
  -- Additional Dimensions
  p.account_id
from
  aws_iam_policy as p
  left join bad_policies as bad on p.arn = bad.arn
where
  p.arn not like 'arn:aws:iam::aws:policy%'
