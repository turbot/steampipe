with backup_protected_file_system as (
  select
    resource_arn as arn
  from
    aws_backup_protected_resource as b
  where
    resource_type = 'EFS'
)
select
  -- Required Columns
  f.arn as resource,
  case
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.arn is not null then f.title || ' is protected by backup plan.'
    else f.title || ' is not protected by backup plan.'
  end as reason,
  -- Additional Dimensions
  f.region,
  f.account_id
from
  aws_efs_file_system as f
  left join backup_protected_file_system as b on f.arn = b.arn;
