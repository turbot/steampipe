with backup_protected_rds_isntance as (
  select
    resource_arn as arn
  from
    aws_backup_protected_resource as b
  where
    resource_type = 'RDS'
)
select
  -- Required Columns
  r.arn as resource,
  case
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.arn is not null then r.title || ' is protected by backup plan.'
    else r.title || ' is not protected by backup plan.'
  end as reason,
  -- Additional Dimensions
  r.region,
  r.account_id
from
  aws_rds_db_instance as r
  left join backup_protected_rds_isntance as b on r.arn = b.arn;
