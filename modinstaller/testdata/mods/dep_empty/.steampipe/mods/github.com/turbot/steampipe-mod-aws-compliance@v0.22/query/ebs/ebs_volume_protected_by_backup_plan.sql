with backup_protected_volume as (
  select
    resource_arn as arn
  from
    aws_backup_protected_resource as b
  where
    resource_type = 'EBS'
)
select
  -- Required Columns
  v.arn as resource,
  case
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.arn is not null then v.title || ' is protected by backup plan.'
    else v.title || ' is not protected by backup plan.'
  end as reason,
  -- Additional Dimensions
  v.region,
  v.account_id
from
  aws_ebs_volume as v
  left join backup_protected_volume as b on v.arn = b.arn;