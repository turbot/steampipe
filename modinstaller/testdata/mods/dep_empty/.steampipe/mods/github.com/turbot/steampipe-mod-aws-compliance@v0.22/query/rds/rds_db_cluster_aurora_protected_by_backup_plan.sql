with backup_protected_cluster as (
  select
    resource_arn as arn
  from
    aws_backup_protected_resource as b
  where
    resource_type = 'Aurora'
)
select
  -- Required Columns
  c.arn as resource,
  case
    when c.engine not like '%aurora%' then 'skip'
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when c.engine not like '%aurora%' then c.title || ' not Aurora resources.'
    when b.arn is not null then c.title || ' is protected by backup plan.'
    else c.title || ' is not protected by backup plan.'
  end as reason,
  -- Additional Dimensions
  c.region,
  c.account_id
from
  aws_rds_db_cluster as c
  left join backup_protected_cluster as b on c.arn = b.arn;