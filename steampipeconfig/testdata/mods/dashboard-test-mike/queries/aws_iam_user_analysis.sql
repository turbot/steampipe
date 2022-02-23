with args as (
    select 'arn:aws:iam::876515858155:user/mike' as iam_user_arn
),
analysis as (
  select
    null as parent,
    arn as id,
    title as name,
    0 as depth,
    'aws_iam_user' as category
  from
    aws_iam_user
  where
    arn in (select iam_user_arn from args)
  union
  select
    u.arn as parent,
    ak.access_key_id as id,
    ak.title as name,
    1 as depth,
    'aws_iam_access_key' as category
  from
    aws_iam_access_key ak
    inner join aws_iam_user u on ak.user_name = u.name
  where
    u.arn in (select iam_user_arn from args)
  union
  select
    u.arn as parent,
    g.arn as id,
    g.title as name,
    1 as depth,
    'aws_iam_group' as category
  from
    aws_iam_user u,
    jsonb_array_elements(u.groups) as user_groups
    inner join aws_iam_group g on g.arn = user_groups ->> 'Arn'
  where
    u.arn in (select iam_user_arn from args)
  union
  select
    g.arn as parent,
    p.arn as id,
    p.title as name,
    2 as depth,
    'aws_iam_policy' as category
  from
    aws_iam_user as u,
    aws_iam_policy as p,
    jsonb_array_elements(u.groups) as user_groups
    inner join aws_iam_group g on g.arn = user_groups ->> 'Arn'
  where
    g.attached_policy_arns :: jsonb ? p.arn
    and u.arn in (select iam_user_arn from args)
  union
  select
    u.arn as parent,
    p.arn as id,
    p.title as name,
    2 as depth,
    'aws_iam_policy' as category
  from
    aws_iam_user as u,
    jsonb_array_elements_text(u.attached_policy_arns) as pol_arn,
    aws_iam_policy as p
  where
    u.attached_policy_arns :: jsonb ? p.arn
    and pol_arn = p.arn
    and u.arn in (select iam_user_arn from args)
)
select
  *
from
  analysis
order by
  depth,
  category,
  id;