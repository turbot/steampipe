with backup_protected_instance as (
  select
    resource_arn as arn
  from
    aws_backup_protected_resource as b
  where
    resource_type = 'EC2'
)
select
  -- Required Columns
  i.arn as resource,
  case
    when b.arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when b.arn is not null then i.title || ' is protected by backup plan.'
    else i.title || ' is not protected by backup plan.'
  end as reason,
  -- Additional Dimensions
  i.region,
  i.account_id
from
  aws_ec2_instance as i
  left join backup_protected_instance as b on i.arn = b.arn;
